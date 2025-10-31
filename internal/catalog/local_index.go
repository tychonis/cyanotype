package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tychonis/cyanotype/model"
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

type Index interface {
	Index(sym model.ConcreteSymbol) error

	Find(q Qualifier) (model.Digest, error)
	GetType(digest model.Digest) (string, error)
	GetItemProcesses(item model.ItemID) ([]model.ProcessID, error)
	GetItemCoProcesses(item model.ItemID) ([]model.ProcessID, error)

	ListSymbols() (map[model.Digest]string, error)
}

type LocalIndex struct {
	qualifierIndex map[Qualifier]model.ItemID
	processIndex   map[model.ItemID]*ProcessIndexEntry
	typeIndex      map[model.Digest]string

	persistent bool
}

func NewLocalIndex(persistent bool) *LocalIndex {
	idx := &LocalIndex{
		qualifierIndex: make(map[Qualifier]model.Digest),
		processIndex:   make(map[Qualifier]*ProcessIndexEntry),
		typeIndex:      make(map[model.Digest]string),

		persistent: persistent,
	}
	idx.load()
	return idx
}

func (idx *LocalIndex) load() error {
	err := idx.loadMainIndex()
	if err != nil {
		return err
	}
	err = idx.loadProcessIndex()
	if err != nil {
		return err
	}
	return idx.loadTypeIndex()
}

func (idx *LocalIndex) loadMainIndex() error {
	if !idx.persistent {
		return nil
	}

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

		idx.qualifierIndex[qualifier] = dhex
	}
	return nil
}

func (idx *LocalIndex) addToMainIndex(key string, val string) error {
	oldVal, ok := idx.qualifierIndex[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	idx.qualifierIndex[key] = val

	if !idx.persistent {
		return nil
	}

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

func (idx *LocalIndex) loadTypeIndex() error {
	if !idx.persistent {
		return nil
	}

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

		idx.typeIndex[key] = val
	}
	return nil
}

func (idx *LocalIndex) addToTypeIndex(key string, val string) error {
	oldVal, ok := idx.typeIndex[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	idx.typeIndex[key] = val

	if !idx.persistent {
		return nil
	}

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

func (idx *LocalIndex) loadProcessIndex() error {
	if !idx.persistent {
		return nil
	}

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

		entry, ok := idx.processIndex[key]
		if !ok || entry == nil {
			entry = NewProcessIndexEntry()
			idx.processIndex[key] = entry
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

func (idx *LocalIndex) addToProcessIndex(pType string, key string, val string) error {
	entry, ok := idx.processIndex[key]
	if !ok || entry == nil {
		entry = NewProcessIndexEntry()
		idx.processIndex[key] = entry
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

	if !idx.persistent {
		return nil
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

func (idx *LocalIndex) indexProcess(sym model.ConcreteSymbol) error {
	switch resolved := sym.(type) {
	case *model.Process:
		for _, bomLine := range resolved.Input {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
	case *model.CoProcess:
		for _, bomLine := range resolved.Input {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

// TODO: fix this hack.
func (idx *LocalIndex) indexType(sym model.ConcreteSymbol) error {
	switch sym.(type) {
	case *model.Process:
		idx.addToTypeIndex(sym.GetDigest(), "process")
	case *model.CoProcess:
		idx.addToTypeIndex(sym.GetDigest(), "coprocess")
	case *model.Item:
		idx.addToTypeIndex(sym.GetDigest(), "item")
	case *model.CoItem:
		idx.addToTypeIndex(sym.GetDigest(), "coitem")
	}
	return nil
}

func (idx *LocalIndex) Index(sym model.ConcreteSymbol) error {
	err := idx.addToMainIndex(sym.GetQualifier(), sym.GetDigest())
	if err != nil {
		return err
	}
	err = idx.indexProcess(sym)
	if err != nil {
		return err
	}
	return idx.indexType(sym)
}

func (idx *LocalIndex) Find(q Qualifier) (model.Digest, error) {
	digest, ok := idx.qualifierIndex[q]
	if !ok {
		return "", ErrNotFound
	}
	return digest, nil
}

func (idx *LocalIndex) GetType(digest model.Digest) (string, error) {
	t, ok := idx.typeIndex[digest]
	if !ok {
		return "", ErrNotFound
	}
	return t, nil
}

func (idx *LocalIndex) GetItemProcesses(item model.ItemID) ([]model.ProcessID, error) {
	entry, ok := idx.processIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.Processes, nil
}

func (idx *LocalIndex) GetItemCoProcesses(item model.ItemID) ([]model.ProcessID, error) {
	entry, ok := idx.processIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.CoProcesses, nil
}

func (idx *LocalIndex) ListSymbols() (map[model.Digest]string, error) {
	return idx.typeIndex, nil
}
