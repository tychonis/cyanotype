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

type SymbolIndex interface {
	IndexSymbol(r *model.Revision, sym model.ConcreteSymbol) error

	FindAll(q Qualifier) ([]model.Digest, error)
	FindCurrent(q Qualifier) (model.Digest, error)

	GetItemProcesses(item model.ItemID) ([]process.ProcessID, error)
	GetItemCoProcesses(item model.ItemID) ([]process.ProcessID, error)
}

type RevisionIndex interface {
	IndexRevision(r *model.Revision) error
	GetRevision(r model.RevisionID) (*model.Revision, error)
	CompareRevisions(r1, r2 model.RevisionID) int
}

type Index interface {
	SymbolIndex
	RevisionIndex
}
