package ranker

import (
	"fmt"
	"sort"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/process"
)

type CatalogRanker struct {
	catalog *catalog.Catalog
}

func (r *CatalogRanker) GetProcessRank(p *process.Process) (*catalog.Rank, error) {
	sym, err := r.catalog.Get(p.Digest)
	if err != nil {
		return catalog.ZeroRank(), err
	}
	p, ok := sym.(*process.Process)
	if !ok {
		return catalog.ZeroRank(), fmt.Errorf("symbol is not a process")
	}
	metadata, err := r.catalog.GetMetadata(p.Digest)
	if err != nil {
		return catalog.ZeroRank(), err
	}
	return metadata.Rank, nil
}

func (r *CatalogRanker) GetCoProcessRank(cp *process.CoProcess) (*catalog.Rank, error) {
	sym, err := r.catalog.Get(cp.Digest)
	if err != nil {
		return catalog.ZeroRank(), err
	}
	cp, ok := sym.(*process.CoProcess)
	if !ok {
		return catalog.ZeroRank(), fmt.Errorf("symbol is not a coprocess")
	}
	metadata, err := r.catalog.GetMetadata(cp.Digest)
	if err != nil {
		return catalog.ZeroRank(), err
	}
	return metadata.Rank, nil
}

func (r *CatalogRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	sort.SliceStable(cps, func(i, j int) bool {
		rankI, errI := r.GetCoProcessRank(cps[i])
		rankJ, errJ := r.GetCoProcessRank(cps[j])

		if errI != nil || errJ != nil {
			return false
		}

		return catalog.CmpRank(*rankI, *rankJ) < 0
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

		return catalog.CmpRank(*rankI, *rankJ) < 0
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
