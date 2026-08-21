package catalog

import (
	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

type Qualifier = string

type ProcessIndexEntry struct {
	Processes   []process.ProcessID
	CoProcesses []process.ProcessID
}

func NewProcessIndexEntry() *ProcessIndexEntry {
	return &ProcessIndexEntry{
		Processes:   make([]process.ProcessID, 0),
		CoProcesses: make([]process.ProcessID, 0),
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

	GetItemProcesses(item model.ItemID) ([]process.ProcessID, error)
	GetItemCoProcesses(item model.ItemID) ([]process.ProcessID, error)

	GetContent() *IndexContent
}

type RevisionIndex interface {
	IndexRevision(r *model.Revision) error
	// GetRevision(r model.RevisionID) (*model.Revision, error)
	// CompareRevisions return negative if a is older than b.
	CompareRevisions(r1, r2 model.RevisionID) int
	GetAllRevisions() ([]model.RevisionID, error)
	GetLatestRevision() (*model.Revision, error)
	GetNewerRevisions(r model.RevisionID) ([]model.RevisionID, error)
}

type Index interface {
	SymbolIndex
	RevisionIndex
}
