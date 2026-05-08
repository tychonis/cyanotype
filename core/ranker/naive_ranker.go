package ranker

import "github.com/tychonis/cyanotype/core/process"

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
