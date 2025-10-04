package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model/v2"
)

type Qualifier = string

type ProcessIndexEntry struct {
	Processes   []model.ProcessID
	CoProcesses []model.ProcessID
}

func NewProcessIndexEntry() *ProcessIndexEntry {
	return &ProcessIndexEntry{
		Processes:   make([]Qualifier, 0),
		CoProcesses: make([]Qualifier, 0),
	}
}

type ItemProcess = struct {
	Item    model.ItemID
	Process model.ProcessID
}

var ErrNotFound = errors.New("symbol not found")

type Catalog interface {
	Add(symbol model.ConcreteSymbol) error
	Get(digest model.Digest) (model.ConcreteSymbol, error)
	Find(qualifier Qualifier) (model.ConcreteSymbol, error)

	GetItemProcesses(item model.ItemID) ([]*model.Process, error)
	GetItemCoProcesses(item model.ItemID) ([]*model.CoProcess, error)

	GetCoItems(item model.ItemID) ([]*ItemProcess, error)
	GetItems(coItem model.ItemID) ([]*ItemProcess, error)
}

type LocalCatalog struct {
	index        map[Qualifier]model.ItemID
	processIndex map[model.ItemID]*ProcessIndexEntry
	typeIndex    map[model.ItemID]string
}

func NewLocalCatalog() *LocalCatalog {
	cat := &LocalCatalog{
		index:        make(map[Qualifier]model.Digest),
		processIndex: make(map[Qualifier]*ProcessIndexEntry),
		typeIndex:    make(map[model.Digest]string),
	}
	cat.loadMainIndex()
	cat.loadProcessIndex()
	cat.loadTypeIndex()
	return cat
}

func (c *LocalCatalog) loadMainIndex() error {
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

		qualifier := string(parts[0])
		dhex := string(parts[1])

		c.index[qualifier] = dhex
	}
	return nil
}

func (c *LocalCatalog) addToMainIndex(key string, val string) error {
	oldVal, ok := c.index[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	c.index[key] = val

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

func (c *LocalCatalog) loadTypeIndex() error {
	indexPath := filepath.Join(".bpc", "types")
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

		key := string(parts[0])
		val := string(parts[1])

		c.typeIndex[key] = val
	}
	return nil
}

func (c *LocalCatalog) addToTypeIndex(key string, val string) error {
	oldVal, ok := c.typeIndex[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	c.typeIndex[key] = val

	indexPath := filepath.Join(".bpc", "types")
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

func (c *LocalCatalog) loadProcessIndex() error {
	indexPath := filepath.Join(".bpc", "process")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte(":"), 3)
		if len(parts) != 3 {
			return fmt.Errorf("malformed part")
		}

		pType := string(parts[0])
		key := string(parts[1])
		val := string(parts[2])

		entry, ok := c.processIndex[key]
		if !ok || entry == nil {
			entry = NewProcessIndexEntry()
			c.processIndex[key] = entry
		}
		switch pType {
		case "process":
			for _, p := range entry.Processes {
				if p == val {
					return nil
				}
			}
			entry.Processes = append(entry.Processes, val)
		case "coprocess":
			for _, p := range entry.CoProcesses {
				if p == val {
					return nil
				}
			}
			entry.CoProcesses = append(entry.CoProcesses, val)
		default:
			return errors.New("illegal process type")
		}
	}
	return nil
}

func (c *LocalCatalog) addToProcessIndex(pType string, key string, val string) error {
	entry, ok := c.processIndex[key]
	if !ok || entry == nil {
		entry = NewProcessIndexEntry()
		c.processIndex[key] = entry
	}

	switch pType {
	case "process":
		for _, p := range entry.Processes {
			if p == val {
				return nil
			}
		}
		entry.Processes = append(entry.Processes, val)
	case "coprocess":
		for _, p := range entry.CoProcesses {
			if p == val {
				return nil
			}
		}
		entry.CoProcesses = append(entry.CoProcesses, val)
	default:
		return errors.New("illegal process type")
	}

	indexPath := filepath.Join(".bpc", "process")
	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	defer f.Close()
	rec := pType + ":" + key + ":" + val + "\n"
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

func (c *LocalCatalog) indexProcess(sym model.ConcreteSymbol) error {
	switch resolved := sym.(type) {
	case *model.Process:
		for _, bomLine := range resolved.Input {
			c.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			c.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
	case *model.CoProcess:
		for _, bomLine := range resolved.Input {
			c.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			c.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

// TODO: fix this hack.
func (c *LocalCatalog) indexType(sym model.ConcreteSymbol) error {
	switch sym.(type) {
	case *model.Process:
		c.addToTypeIndex(sym.GetDigest(), "process")
	case *model.CoProcess:
		c.addToTypeIndex(sym.GetDigest(), "coprocess")
	case *model.Item:
		c.addToTypeIndex(sym.GetDigest(), "item")
	case *model.CoItem:
		c.addToTypeIndex(sym.GetDigest(), "coitem")
	}
	return nil
}

func (c *LocalCatalog) Add(sym model.ConcreteSymbol) error {
	body, err := serializer.Serialize(sym)
	if err != nil {
		return err
	}
	c.addToMainIndex(sym.GetQualifier(), sym.GetDigest())
	c.indexProcess(sym)
	c.indexType(sym)
	return atomicWrite(digestToPath(sym.GetDigest()), body, 0o644)
}

func (c *LocalCatalog) Get(digest model.Digest) (model.ConcreteSymbol, error) {
	path := digestToPath(digest)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t, ok := c.typeIndex[digest]
	if !ok {
		return nil, errors.New("error deciding type")
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

func (c *LocalCatalog) Find(qualifier Qualifier) (model.ConcreteSymbol, error) {
	digest, ok := c.index[qualifier]
	if !ok {
		return nil, ErrNotFound
	}
	return c.Get(digest)
}

func getSymbols[T model.ConcreteSymbol](c Catalog, ids []model.Digest) ([]T, error) {
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

func (c *LocalCatalog) GetItemProcesses(item model.ItemID) ([]*model.Process, error) {
	index, ok := c.processIndex[item]
	if !ok {
		return nil, nil
	}
	return getSymbols[*model.Process](c, index.Processes)
}

func (c *LocalCatalog) GetItemCoProcesses(item model.ItemID) ([]*model.CoProcess, error) {
	index, ok := c.processIndex[item]
	if !ok {
		slog.Debug("nocoprocess found", "item", item)
		return nil, nil
	}
	slog.Debug("found coprocess", "item", item)
	return getSymbols[*model.CoProcess](c, index.CoProcesses)
}

func (c *LocalCatalog) GetCoItems(item model.ItemID) ([]*ItemProcess, error) {
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

func (c *LocalCatalog) GetItems(coItem model.ItemID) ([]*ItemProcess, error) {
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
