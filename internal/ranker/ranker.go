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

type NaiveRanker struct{}

func (r *NaiveRanker) RankCoProcess(cps []*model.CoProcess) ([]*model.CoProcess, error) {
	return cps, nil
}

func (r *NaiveRanker) TopCoProcess(cps []*model.CoProcess) (*model.CoProcess, error) {
	if len(cps) > 0 {
		return cps[0], nil
	}
	return nil, ErrNotFound
}

func (r *NaiveRanker) RankProcess(ps []*model.Process) ([]*model.Process, error) {
	return ps, nil
}

func (r *NaiveRanker) TopProcess(ps []*model.Process) (*model.Process, error) {
	if len(ps) > 0 {
		return ps[0], nil
	}
	return nil, ErrNotFound
}
