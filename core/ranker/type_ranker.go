package ranker

import "github.com/tychonis/cyanotype/core/process"

type TypeRanker struct {
	PreferedType string
}

func (r *TypeRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	return cps, nil
}

func (r *TypeRanker) TopCoProcess(cps []*process.CoProcess) (*process.CoProcess, error) {
	if len(cps) > 0 {
		return cps[0], nil
	}
	return nil, ErrNotFound
}

func (r *TypeRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
	ret := make([]*process.Process, 0, len(ps))

	for _, p := range ps {
		if p != nil && p.Content != nil && p.Content.GetType() == r.PreferedType {
			ret = append(ret, p)
		}
	}

	for _, p := range ps {
		if p == nil || p.Content == nil || p.Content.GetType() != r.PreferedType {
			ret = append(ret, p)
		}
	}

	return ret, nil
}

func (r *TypeRanker) TopProcess(ps []*process.Process) (*process.Process, error) {
	for _, p := range ps {
		if p.Content.GetType() == r.PreferedType {
			return p, nil
		}
	}
	if len(ps) > 0 {
		return ps[0], nil
	}
	return nil, ErrNotFound
}
