package catalog

import (
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model"
)

type ItemProcess = struct {
	Item    model.ItemID
	Process process.ProcessID
}

var ErrNotFound = errors.New("symbol not found")

type Catalog struct {
	storage Storage
	index   Index

	sequence int
}

func NewCatalog(catalogType string) *Catalog {
	switch catalogType {
	case "memory":
		return NewMemoryCatalog()
	default:
		return NewLocalCatalog()
	}
}

func NewLocalCatalog() *Catalog {
	return &Catalog{
		storage: &LocalStorage{},
		index:   NewLocalIndex(true),
	}
}

func NewMemoryCatalog() *Catalog {
	cat := &Catalog{
		storage: NewMemoryStore(),
		index:   NewLocalIndex(false),
	}
	return cat
}

// Adhoc hardcoded remote catalog.
func NewRemoteCatalog(endpoint string, token string, tag string) *Catalog {
	client := NewClient(token)
	cat := &Catalog{
		storage: NewAPIStore(endpoint+"/definition", client),
		index:   NewRemoteIndex(endpoint+"/bom_index/"+tag, client),
	}
	return cat
}

func (c *Catalog) UpdateSymbol(old model.ConcreteSymbol, new model.ConcreteSymbol) error {
	switch oldSym := old.(type) {
	case *model.Item:
		newSym, ok := new.(*model.Item)
		if !ok {
			return errors.New("type mismatch: expected Item")
		}
		coProcesses, err := c.GetItemCoProcesses(oldSym.Digest)
		if err != nil {
			return err
		}
		for _, cp := range coProcesses {
			newCp := *cp
			for _, bomLine := range newCp.Input() {
				if bomLine.Item == oldSym.Digest {
					bomLine.Item = newSym.Digest
				}
			}
			err := c.Add(&newCp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Catalog) GetRank() *Rank {
	defer func() {
		c.sequence += 1
	}()
	return &Rank{
		Sequence: c.sequence,
		WallTime: time.Now().UnixNano(),
	}
}

func (c *Catalog) GenerateMetadata(sym model.ConcreteSymbol) *Metadata {
	return &Metadata{
		Rank: c.GetRank(),
	}
}

func (c *Catalog) Add(sym model.ConcreteSymbol) error {
	oldSym, err := c.Find(sym.GetQualifier())
	if err == nil {
		if oldSym.GetDigest() == sym.GetDigest() {
			return nil
		}
	}

	err = c.index.Index(sym)
	if err != nil {
		return err
	}
	metadata := c.GenerateMetadata(sym)
	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	err = c.storage.SaveMetadata(sym.GetDigest(), body)
	if err != nil {
		return err
	}
	body, err = serializer.Serialize(sym)
	if err != nil {
		return err
	}
	return c.storage.Save(sym.GetDigest(), body)
}

func (c *Catalog) Get(digest model.Digest) (model.ConcreteSymbol, error) {
	body, err := c.storage.Load(digest)
	if err != nil {
		return nil, err
	}
	t, err := c.index.GetType(digest)
	if err != nil {
		return nil, err
	}
	switch t {
	case "item":
		ret, err := serializer.Deserialize[*model.Item](body)
		if err != nil {
			return ret, err
		}
		ret.Digest = digest
		return ret, nil
	case "coitem":
		ret, err := serializer.Deserialize[*model.CoItem](body)
		if err != nil {
			return ret, err
		}
		ret.Digest = digest
		return ret, nil
	case "process":
		ret, err := serializer.Deserialize[*process.Process](body)
		if err != nil {
			return ret, err
		}
		ret.Digest = digest
		return ret, nil
	case "coprocess":
		ret, err := serializer.Deserialize[*process.CoProcess](body)
		if err != nil {
			return ret, err
		}
		ret.Digest = digest
		return ret, nil
	default:
		return nil, errors.New("unknown type")
	}
}

func (c *Catalog) Find(qualifier Qualifier) (model.ConcreteSymbol, error) {
	digest, err := c.index.Find(qualifier)
	if err != nil {
		return nil, err
	}
	return c.Get(digest)
}

func getSymbols[T model.ConcreteSymbol](c *Catalog, ids []model.Digest) ([]T, error) {
	ret := make([]T, 0, len(ids))
	for _, pid := range ids {
		sym, err := c.Get(pid)
		if err != nil {
			return nil, err
		}
		s, ok := sym.(T)
		if !ok {
			slog.Warn("Unexpected symbol type",
				"expected", reflect.TypeOf(ret),
				"pid", pid,
				"type", reflect.TypeOf(sym),
			)
			return nil, errors.New("incorrect symbol type")
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (c *Catalog) GetItemProcesses(item model.ItemID) ([]*process.Process, error) {
	processes, err := c.index.GetItemProcesses(item)
	if err != nil {
		return nil, err
	}
	return getSymbols[*process.Process](c, processes)
}

func (c *Catalog) GetItemCoProcesses(item model.ItemID) ([]*process.CoProcess, error) {
	coProcesses, err := c.index.GetItemCoProcesses(item)
	if err != nil {
		slog.Debug("nocoprocess found", "item", item)
		return nil, err
	}
	slog.Debug("found coprocess", "item", item)
	return getSymbols[*process.CoProcess](c, coProcesses)
}

func (c *Catalog) GetCoItems(item model.ItemID) ([]*ItemProcess, error) {
	slog.Debug("get coitems", "item", item)
	cps, err := c.GetItemCoProcesses(item)
	if err != nil {
		return nil, err
	}
	ret := make([]*ItemProcess, 0, len(cps))
	for _, cp := range cps {
		slog.Debug("coprocess", "item", item, "coprocess", cp.Digest)
		if len(cp.Output()) != 1 {
			return nil, errors.New("multiple output not implemented yet")
		}
		ret = append(ret, &ItemProcess{
			Item:    cp.Output()[0].Item,
			Process: cp.Digest,
		})
	}
	return ret, nil
}

func (c *Catalog) GetItems(coItem model.ItemID) ([]*ItemProcess, error) {
	cps, err := c.GetItemCoProcesses(coItem)
	if err != nil {
		return nil, err
	}
	ret := make([]*ItemProcess, 0, len(cps))
	for _, cp := range cps {
		if len(cp.Output()) != 1 {
			return nil, errors.New("multiple input not implemented yet")
		}
		ret = append(ret, &ItemProcess{
			Item:    cp.Input()[0].Item,
			Process: cp.Digest,
		})
	}
	return ret, nil
}

func (c *Catalog) GetMetadata(digest model.Digest) (*Metadata, error) {
	data, err := c.storage.LoadMetadata(digest)
	if err != nil {
		return nil, err
	}
	metadata := &Metadata{}
	err = json.Unmarshal(data, metadata)
	return metadata, err
}
