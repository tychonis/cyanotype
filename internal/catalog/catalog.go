package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model/v2"
)

type Qualifier = string
type Digest = string

type IndexEntry struct {
	Processes   []Digest
	CoProcesses []Digest
}

func NewIndexEntry() *IndexEntry {
	return &IndexEntry{
		Processes:   make([]Qualifier, 0),
		CoProcesses: make([]Qualifier, 0),
	}
}

type Catalog interface {
	Add(symbol model.ConcreteSymbol) error
	Get(digest Digest) (model.ConcreteSymbol, error)
	Find(qualifier Qualifier) (model.ConcreteSymbol, error)

	GetItemProcesses(item Digest) ([]*model.Process, error)
	GetItemCoProcesses(item Digest) ([]*model.CoProcess, error)
}

type LocalCatalog struct {
	index        map[Qualifier]Digest
	processIndex map[model.ItemID]*IndexEntry
}

func NewLocalCatalog() *LocalCatalog {
	cat := &LocalCatalog{
		index:        make(map[Qualifier]Digest),
		processIndex: make(map[Qualifier]*IndexEntry),
	}
	cat.loadIndex()
	return cat
}

func (c *LocalCatalog) loadIndex() error {
	indexPath := filepath.Join(".bpc", "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return fmt.Errorf("malformed part")
		}

		dhex := string(parts[0])
		qualifier := string(parts[1])

		c.index[dhex] = qualifier
	}
	return nil
}

func (c *LocalCatalog) appendIndexItem(key string, val string) error {
	indexPath := filepath.Join(".bpc", "index")
	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	defer f.Close()
	rec := key + ":" + val + "\n"
	_, err = f.Write([]byte(rec))
	if err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return f.Sync()
}

func digestToPath(digest string) string {
	folder := digest[:2]
	return filepath.Join(".bpc", "objects", folder, digest)
}

func (c *LocalCatalog) linkProcessToItem(item model.ItemID, process model.ProcessID) {
	if c.processIndex[item] == nil {
		c.processIndex[item] = NewIndexEntry()
	}
	c.processIndex[item].Processes = append(c.processIndex[item].Processes, process)
}

func (c *LocalCatalog) linkCoProcessToItem(item model.ItemID, coProcess model.ProcessID) {
	if c.processIndex[item] == nil {
		c.processIndex[item] = NewIndexEntry()
	}
	c.processIndex[item].Processes = append(c.processIndex[item].CoProcesses, coProcess)
}

func (c *LocalCatalog) indexProcess(sym model.ConcreteSymbol) error {
	switch resolved := sym.(type) {
	case *model.Process:
		for _, bomLine := range resolved.Input {
			c.linkProcessToItem(bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			c.linkProcessToItem(bomLine.Item, resolved.Digest)
		}
	case *model.CoProcess:
		for _, bomLine := range resolved.Input {
			c.linkCoProcessToItem(bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			c.linkCoProcessToItem(bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

func (c *LocalCatalog) Add(sym model.ConcreteSymbol) error {
	body, err := serializer.Serialize(sym)
	if err != nil {
		return err
	}
	c.index[sym.GetQualifier()] = sym.GetDigest()
	c.appendIndexItem(sym.GetQualifier(), sym.GetDigest())
	c.indexProcess(sym)
	return atomicWrite(digestToPath(sym.GetDigest()), body, 0o644)
}

func (c *LocalCatalog) Get(digest Digest) (model.ConcreteSymbol, error) {
	path := digestToPath(digest)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ret, err := serializer.Deserialize[*model.Item](body)
	if err != nil {
		return ret, err
	}
	ret.Digest = digest
	return ret, nil
}

func (c *LocalCatalog) Find(qualifier Qualifier) (model.ConcreteSymbol, error) {
	digest, ok := c.index[qualifier]
	if !ok {
		return nil, fmt.Errorf("could not find qualifier %s", qualifier)
	}
	return c.Get(digest)
}

func getSymbols[T model.ConcreteSymbol](c Catalog, ids []Digest) ([]T, error) {
	ret := make([]T, 0, len(ids))
	for _, pid := range ids {
		sym, err := c.Get(pid)
		if err != nil {
			return nil, err
		}
		s, ok := sym.(T)
		if !ok {
			return nil, errors.New("incorrect symbol type")
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (c *LocalCatalog) GetItemProcesses(item Digest) ([]*model.Process, error) {
	index, ok := c.processIndex[item]
	if !ok {
		return nil, nil
	}
	return getSymbols[*model.Process](c, index.Processes)
}

func (c *LocalCatalog) GetItemCoProcesses(item Digest) ([]*model.CoProcess, error) {
	index, ok := c.processIndex[item]
	if !ok {
		return nil, nil
	}
	return getSymbols[*model.CoProcess](c, index.CoProcesses)
}
