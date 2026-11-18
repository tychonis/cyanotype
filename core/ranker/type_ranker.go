package ranker

import (
	"github.com/tychonis/cyanotype/model"
)

type TypeRanker struct {
	PreferedType string
}

func (r *TypeRanker) RankCoProcess(cps []*model.CoProcess) ([]*model.CoProcess, error) {
	return cps, nil
}

func (r *TypeRanker) TopCoProcess(cps []*model.CoProcess) (*model.CoProcess, error) {
	if len(cps) > 0 {
		return cps[0], nil
	}
	return nil, ErrNotFound
}

func (r *TypeRanker) RankProcess(ps []*model.Process) ([]*model.Process, error) {
	ret := make([]*model.Process, 0, len(ps))

	for _, p := range ps {
		if p != nil && p.Get("purpose") == r.PreferedType {
			ret = append(ret, p)
		}
	}

	for _, p := range ps {
		if p == nil || p.Get("purpose") != r.PreferedType {
			ret = append(ret, p)
		}
	}

	return ret, nil
}

func (r *TypeRanker) TopProcess(ps []*model.Process) (*model.Process, error) {
	for _, p := range ps {
		if p.Get("purpose") == r.PreferedType {
			return p, nil
		}
	}
	if len(ps) > 0 {
		return ps[0], nil
	}
	return nil, ErrNotFound
}
