package ranker

import "github.com/tychonis/cyanotype/model"

type Ranker interface {
	RankCoProcess([]*model.CoProcess) []*model.CoProcess
	TopCoProcess([]*model.CoProcess) *model.CoProcess
	RankProcess([]*model.Process) []*model.Process
	TopProcess([]*model.Process) *model.Process
}

type NaiveRanker struct{}

func (r *NaiveRanker) RankCoProcess(cps []*model.CoProcess) []*model.CoProcess {
	return cps
}

func (r *NaiveRanker) TopCoProcess(cps []*model.CoProcess) *model.CoProcess {
	if len(cps) > 0 {
		return cps[0]
	}
	return nil
}

func (r *NaiveRanker) RankProcess(ps []*model.Process) []*model.Process {
	return ps
}

func (r *NaiveRanker) TopProcess(ps []*model.Process) *model.Process {
	if len(ps) > 0 {
		return ps[0]
	}
	return nil
}
