package ranker

import (
	"errors"

	"github.com/tychonis/cyanotype/model"
)

var ErrNotFound = errors.New("not found")

type Ranker interface {
	RankCoProcess([]*model.CoProcess) ([]*model.CoProcess, error)
	TopCoProcess([]*model.CoProcess) (*model.CoProcess, error)
	RankProcess([]*model.Process) ([]*model.Process, error)
	TopProcess([]*model.Process) (*model.Process, error)
}
