package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

type LocalIndex struct {
	qualifierIndex map[Qualifier]QualifierIndexEntry
	processIndex   map[model.ItemID]*ProcessIndexEntry
	revisionIndex  map[model.RevisionID]*model.Revision

	persistent bool
}

func NewLocalIndex(persistent bool) *LocalIndex {
	idx := &LocalIndex{
		qualifierIndex: make(map[Qualifier]QualifierIndexEntry),
		processIndex:   make(map[model.ItemID]*ProcessIndexEntry),
		revisionIndex:  make(map[model.RevisionID]*model.Revision),

		persistent: persistent,
	}
	idx.load()
	return idx
}

func (idx *LocalIndex) load() error {
	err := idx.loadQualifierIndex()
	if err != nil {
		return err
	}
	err = idx.loadProcessIndex()
	if err != nil {
		return err
	}
	err = idx.loadRevisionIndex()
	if err != nil {
		return err
	}
	return nil
}

func (idx *LocalIndex) loadQualifierIndex() error {
	if !idx.persistent {
		return nil
	}

	indexPath := filepath.Join(".bpc", "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("loading index error, failed to open index: %w", err)
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

		qualifier := string(parts[0])
		revision := model.RevisionID(parts[1])
		symDigest := model.Digest(parts[2])
		entry, ok := idx.qualifierIndex[qualifier]
		if !ok {
			entry = make(QualifierIndexEntry)
			idx.qualifierIndex[qualifier] = entry
		}
		entry[revision] = symDigest
	}
	return nil
}

func (idx *LocalIndex) addToQualifierIndex(qualifier string, revision model.RevisionID, symDigest model.Digest) error {
	currentEntry, ok := idx.qualifierIndex[qualifier]
	if !ok {
		currentEntry = make(QualifierIndexEntry)
		idx.qualifierIndex[qualifier] = currentEntry
	}
	currentEntry[revision] = symDigest

	if !idx.persistent {
		return nil
	}

	indexPath := filepath.Join(".bpc", "index")
	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to add to index, cannot open index: %w", err)
	}
	defer f.Close()
	rec := qualifier + ":" + string(revision) + ":" + string(symDigest) + "\n"
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
	case *process.Process:
		for _, bomLine := range resolved.Input() {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output() {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
	case *process.CoProcess:
		for _, bomLine := range resolved.Input() {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output() {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

func (idx *LocalIndex) IndexSymbol(rev *model.Revision, sym model.ConcreteSymbol) error {
	err := idx.addToQualifierIndex(sym.GetQualifier(), rev.Digest, sym.GetDigest())
	if err != nil {
		return err
	}
	return idx.indexProcess(sym)
}

func (idx *LocalIndex) FindAll(q Qualifier) ([]model.Digest, error) {
	entry, ok := idx.qualifierIndex[q]
	if !ok {
		return nil, ErrNotFound
	}
	digests := make([]model.Digest, 0, len(entry))
	for _, digest := range entry {
		digests = append(digests, digest)
	}
	return digests, nil
}

func (idx *LocalIndex) GetRevision(r model.RevisionID) (*model.Revision, error) {
	revision, ok := idx.revisionIndex[r]
	if !ok {
		return nil, ErrNotFound
	}
	return revision, nil
}

func (idx *LocalIndex) CompareRevisions(a, b model.RevisionID) int {
	revA, err := idx.GetRevision(a)
	if err != nil {
		return 0
	}
	revB, err := idx.GetRevision(b)
	if err != nil {
		return 0
	}
	// TODO: Compare based on topological order of revisions before CreatedAt.
	if revA.CreatedAt > revB.CreatedAt {
		return -1
	} else if revA.CreatedAt < revB.CreatedAt {
		return 1
	}
	return 0
}

func (idx *LocalIndex) FindCurrent(q Qualifier) (model.Digest, error) {
	entry, ok := idx.qualifierIndex[q]
	if !ok {
		return "", ErrNotFound
	}
	allRevisions := make([]model.RevisionID, 0, len(entry))
	for rev := range entry {
		allRevisions = append(allRevisions, rev)
	}
	if len(allRevisions) == 0 {
		return "", ErrNotFound
	}
	sort.SliceStable(allRevisions, func(i, j int) bool {
		return idx.CompareRevisions(allRevisions[i], allRevisions[j]) < 0
	})
	latestRevision := entry[allRevisions[0]]
	return entry[latestRevision], nil
}

func (idx *LocalIndex) GetItemProcesses(item model.ItemID) ([]process.ProcessID, error) {
	entry, ok := idx.processIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.Processes, nil
}

func (idx *LocalIndex) GetItemCoProcesses(item model.ItemID) ([]process.ProcessID, error) {
	entry, ok := idx.processIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.CoProcesses, nil
}

func (idx *LocalIndex) IndexRevision(r *model.Revision) error {
	idx.revisionIndex[r.Digest] = r
	if idx.persistent {
		indexPath := filepath.Join(".bpc", "revision")
		f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open index: %w", err)
		}
		defer f.Close()
		parents := strings.Join(r.Parents, ",")
		rec := string(r.Digest) + ":" + fmt.Sprint(r.CreatedAt) + ":" + parents + "\n"
		_, err = f.Write([]byte(rec))
		if err != nil {
			return fmt.Errorf("write index: %w", err)
		}
	}
	return nil
}

func (idx *LocalIndex) loadRevisionIndex() error {
	if !idx.persistent {
		return nil
	}

	indexPath := filepath.Join(".bpc", "revision")
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
		if len(parts) < 2 {
			continue
		}
		digest := model.RevisionID(parts[0])
		createdAt, err := strconv.ParseInt(string(parts[1]), 10, 64)
		if err != nil {
			return fmt.Errorf("parse createdAt: %w", err)
		}
		var parents []model.RevisionID
		if len(parts) == 3 && len(parts[2]) > 0 {
			for _, p := range strings.Split(string(parts[2]), ",") {
				if p != "" {
					parents = append(parents, model.RevisionID(p))
				}
			}
		}
		idx.revisionIndex[digest] = &model.Revision{
			Digest:    digest,
			CreatedAt: createdAt,
			Parents:   parents,
		}
	}
	return nil
}
