package catalog

import (
	"github.com/tychonis/cyanotype/model"
)

type Qualifier = string

type ProcessIndexEntry struct {
	Processes   []model.ProcessID
	CoProcesses []model.ProcessID
}

func NewProcessIndexEntry() *ProcessIndexEntry {
	return &ProcessIndexEntry{
		Processes:   make([]model.ProcessID, 0),
		CoProcesses: make([]model.ProcessID, 0),
	}
}

type QualifierIndexEntry = map[model.RevisionID]model.Digest
type DigestIndexEntry = map[model.RevisionID]Qualifier

type SymbolIndex interface {
	IndexSymbol(r *model.Revision, sym model.ConcreteSymbol) error
	GetAllSymbols() ([]model.Digest, error)

	FindAllDigests(q Qualifier) ([]model.Digest, error)
	FindCurrentDigest(q Qualifier) (model.Digest, error)

	FindAllQualifiers(d model.Digest) ([]Qualifier, error)
	FindCurrentQualifier(d model.Digest) (Qualifier, error)

	GetItemProcesses(item model.ItemID) ([]model.ProcessID, error)
	GetItemCoProcesses(item model.ItemID) ([]model.ProcessID, error)

	GetContent() *IndexContent
}

type RevisionIndex interface {
	IndexRevision(r *model.Revision) error
	CompareRevisions(r1, r2 model.RevisionID) int
	GetAllRevisions() ([]model.RevisionID, error)
	GetLatestRevision() (*model.Revision, error)
	GetNewerRevisions(r model.RevisionID) ([]model.RevisionID, error)
}

type MetadataStorage interface {
	GetMetadata(d model.Digest) ([]byte, error)
	SaveMetadata(d model.Digest, m []byte) error
}

type Index interface {
	SymbolIndex
	RevisionIndex
	MetadataStorage
}
