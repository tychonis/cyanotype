package catalog

import (
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/internal/digest"
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

	latestRevision *model.Revision
}

func newRevision(parent *model.Revision) *model.Revision {
	digest, err := digest.RandomSHA256()
	if err != nil {
		slog.Error("failed to generate random SHA256 for revision", "error", err)
		return nil
	}
	if parent == nil {
		return &model.Revision{
			Digest:    digest,
			CreatedAt: time.Now().UnixNano(),
		}
	}
	return &model.Revision{
		Digest:    digest,
		CreatedAt: time.Now().UnixNano(),
		Parents:   []model.RevisionID{parent.Digest},
	}
}

func (c *Catalog) NewRevision() *model.Revision {
	return newRevision(c.latestRevision)
}

func New(catalogType string) *Catalog {
	var c *Catalog
	switch catalogType {
	case "memory":
		c = NewMemoryCatalog()
	default:
		c = NewLocalCatalog()
	}
	return c
}

func NewLocalCatalog() *Catalog {
	idx := NewLocalIndex(true)
	latestRevision, _ := idx.GetLatestRevision()
	return &Catalog{
		storage: &LocalStorage{},
		index:   idx,

		latestRevision: latestRevision,
	}
}

func NewMemoryCatalog() *Catalog {
	cat := &Catalog{
		storage: NewMemoryStore(),
		index:   NewLocalIndex(false),
	}
	return cat
}

func (c *Catalog) GenerateMetadata(revision *model.Revision, sym model.ConcreteSymbol) *Metadata {
	return &Metadata{
		IntroducedBy:  revision.Digest,
		CommitHistory: []model.RevisionID{revision.Digest},
	}
}

func (c *Catalog) Revive(rev *model.Revision, digest model.Digest) error {
	metadata, err := c.GetMetadata(digest)
	if err != nil {
		return err
	}
	metadata.Commit(rev.Digest)
	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return c.storage.SaveMetadata(digest, body)
}

func (c *Catalog) Add(rev *model.Revision, sym model.ConcreteSymbol) error {
	existSym, _ := c.Get(sym.GetDigest())
	if existSym != nil {
		return c.Revive(rev, sym.GetDigest())
	}
	err := c.index.IndexSymbol(rev, sym)
	if err != nil {
		return err
	}
	body, err := serializer.Serialize(sym)
	if err != nil {
		return err
	}
	err = c.storage.Save(sym.GetDigest(), body)
	if err != nil {
		return err
	}
	metadata := c.GenerateMetadata(rev, sym)
	body, err = json.Marshal(metadata)
	if err != nil {
		return err
	}
	return c.storage.SaveMetadata(sym.GetDigest(), body)
}

func (c *Catalog) Get(digest model.Digest) (model.ConcreteSymbol, error) {
	body, err := c.storage.Load(digest)
	if err != nil {
		return nil, err
	}
	symType, err := serializer.GetType(body)
	if err != nil {
		return nil, err
	}
	switch symType {
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
		slog.Warn("Unknown symbol type", "type", symType, "digest", digest)
		return nil, errors.New("unknown type")
	}
}

func (c *Catalog) GetSymbolMetadata(digest model.Digest) (*Metadata, error) {
	body, err := c.storage.LoadMetadata(digest)
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (c *Catalog) FindCurrent(qualifier Qualifier) (model.ConcreteSymbol, error) {
	digest, err := c.index.FindCurrent(qualifier)
	if err != nil {
		return nil, err
	}
	return c.Get(digest)
}

func (c *Catalog) FindAll(qualifier Qualifier) ([]model.ConcreteSymbol, error) {
	digests, err := c.index.FindAll(qualifier)
	if err != nil {
		return nil, err
	}
	ret := make([]model.ConcreteSymbol, 0, len(digests))
	for _, digest := range digests {
		sym, err := c.Get(digest)
		if err != nil {
			return nil, err
		}
		ret = append(ret, sym)
	}
	return ret, nil
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

func (c *Catalog) GetSymbols() (map[model.Digest]model.ConcreteSymbol, error) {
	ret := make(map[model.Digest]model.ConcreteSymbol)
	allSymbols, err := c.index.GetAllSymbols()
	if err != nil {
		return nil, err
	}
	for _, digest := range allSymbols {
		sym, err := c.Get(digest)
		if err != nil {
			return nil, err
		}
		ret[digest] = sym
	}
	return ret, nil
}

func (c *Catalog) Commit(revision *model.Revision) error {
	c.index.IndexRevision(revision)
	body, err := serializer.Serialize(revision)
	if err != nil {
		return err
	}
	err = c.storage.Save(revision.Digest, body)
	if err != nil {
		return err
	}
	c.latestRevision = revision
	return nil
}

func (c *Catalog) GetLatestRevision() (*model.Revision, error) {
	return c.latestRevision, nil
}

func (c *Catalog) CompareRevisions(a, b model.RevisionID) int {
	return c.index.CompareRevisions(a, b)
}
