package ranker

import (
	"fmt"
	"sort"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/internal/catalog"
)

type CatalogRanker struct {
	catalog *catalog.Catalog
}

type Rank struct {
	Sequence int
	WallTime int64
}

func ZeroRank() *Rank {
	return &Rank{Sequence: 0, WallTime: 0}
}

func cmpRank(r1, r2 Rank) int {
	if r1.Sequence != r2.Sequence {
		return r1.Sequence - r2.Sequence
	}
	if r1.WallTime != r2.WallTime {
		if r1.WallTime < r2.WallTime {
			return -1
		}
		return 1
	}
	return 0
}

func (r *CatalogRanker) GetProcessRank(p *process.Process) (*Rank, error) {
	sym, err := r.catalog.Get(p.Digest)
	if err != nil {
		return ZeroRank(), err
	}
	p, ok := sym.(*process.Process)
	if !ok {
		return ZeroRank(), fmt.Errorf("symbol is not a process")
	}
	return ZeroRank(), nil
}

func (r *CatalogRanker) GetCoProcessRank(cp *process.CoProcess) (*Rank, error) {
	sym, err := r.catalog.Get(cp.Digest)
	if err != nil {
		return ZeroRank(), err
	}
	cp, ok := sym.(*process.CoProcess)
	if !ok {
		return ZeroRank(), fmt.Errorf("symbol is not a coprocess")
	}
	return ZeroRank(), nil
}

func (r *CatalogRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	sort.SliceStable(cps, func(i, j int) bool {
		rankI, errI := r.GetCoProcessRank(cps[i])
		rankJ, errJ := r.GetCoProcessRank(cps[j])

		if errI != nil || errJ != nil {
			return false
		}

		return cmpRank(*rankI, *rankJ) < 0
	})
	return cps, nil
}

func (r *CatalogRanker) TopCoProcess(cps []*process.CoProcess) (*process.CoProcess, error) {
	if len(cps) > 0 {
		cps, err := r.RankCoProcess(cps)
		if err != nil {
			return nil, err
		}
		return cps[0], nil
	}
	return nil, ErrNotFound
}

func (r *CatalogRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
	sort.SliceStable(ps, func(i, j int) bool {
		rankI, errI := r.GetProcessRank(ps[i])
		rankJ, errJ := r.GetProcessRank(ps[j])

		if errI != nil || errJ != nil {
			return false
		}

		return cmpRank(*rankI, *rankJ) < 0
	})
	return ps, nil
}

func (r *CatalogRanker) TopProcess(ps []*process.Process) (*process.Process, error) {
	if len(ps) > 0 {
		ps, err := r.RankProcess(ps)
		if err != nil {
			return nil, err
		}
		return ps[0], nil
	}
	return nil, ErrNotFound
}
