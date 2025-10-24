package catalog

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model"
)

type ItemProcess = struct {
	Item    model.ItemID
	Process model.ProcessID
}

var ErrNotFound = errors.New("symbol not found")

type Catalog struct {
	storage Storage
	index   Index
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

func (c *Catalog) Add(sym model.ConcreteSymbol) error {
	body, err := serializer.Serialize(sym)
	if err != nil {
		return err
	}
	err = c.index.Index(sym)
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
		ret, err := serializer.Deserialize[*model.Process](body)
		if err != nil {
			return ret, err
		}
		ret.Digest = digest
		return ret, nil
	case "coprocess":
		ret, err := serializer.Deserialize[*model.CoProcess](body)
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
			slog.Debug("incorrect", "pid", pid, "type", reflect.TypeOf(sym))
			return nil, errors.New("incorrect symbol type")
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (c *Catalog) GetItemProcesses(item model.ItemID) ([]*model.Process, error) {
	processes, err := c.index.GetItemProcesses(item)
	if err != nil {
		return nil, err
	}
	return getSymbols[*model.Process](c, processes)
}

func (c *Catalog) GetItemCoProcesses(item model.ItemID) ([]*model.CoProcess, error) {
	coProcesses, err := c.index.GetItemProcesses(item)
	if err != nil {
		slog.Debug("nocoprocess found", "item", item)
		return nil, err
	}
	slog.Debug("found coprocess", "item", item)
	return getSymbols[*model.CoProcess](c, coProcesses)
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
		if len(cp.Output) != 1 {
			return nil, errors.New("multiple output not implemented yet")
		}
		ret = append(ret, &ItemProcess{
			Item:    cp.Output[0].Item,
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
		if len(cp.Output) != 1 {
			return nil, errors.New("multiple input not implemented yet")
		}
		ret = append(ret, &ItemProcess{
			Item:    cp.Input[0].Item,
			Process: cp.Digest,
		})
	}
	return ret, nil
}
