package catalog

import (
	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

type BuildEnv interface {
	Add(rev *model.Revision, sym model.ConcreteSymbol) error
	Get(digest model.Digest) (model.ConcreteSymbol, error)
	FindCurrent(qualifier Qualifier) (model.ConcreteSymbol, error)

	GetItemProcesses(item model.ItemID) ([]*process.Process, error)
	GetItemCoProcesses(item model.ItemID) ([]*process.CoProcess, error)

	Export() ([]byte, error)
}
