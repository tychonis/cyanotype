package ranker

import (
	"errors"

	"github.com/tychonis/cyanotype/core/process"
)

var ErrNotFound = errors.New("not found")

type Ranker interface {
	RankCoProcess([]*process.CoProcess) ([]*process.CoProcess, error)
	TopCoProcess([]*process.CoProcess) (*process.CoProcess, error)
	RankProcess([]*process.Process) ([]*process.Process, error)
	TopProcess([]*process.Process) (*process.Process, error)
}
