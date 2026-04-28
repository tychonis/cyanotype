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

type NaiveRanker struct{}

func (r *NaiveRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	return cps, nil
}

func (r *NaiveRanker) TopCoProcess(cps []*process.CoProcess) (*process.CoProcess, error) {
	if len(cps) > 0 {
		return cps[0], nil
	}
	return nil, ErrNotFound
}

func (r *NaiveRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
	return ps, nil
}

func (r *NaiveRanker) TopProcess(ps []*process.Process) (*process.Process, error) {
	if len(ps) > 0 {
		return ps[0], nil
	}
	return nil, ErrNotFound
}
