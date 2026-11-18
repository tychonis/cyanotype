package ranker

import (
	"github.com/tychonis/cyanotype/model"
)

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
