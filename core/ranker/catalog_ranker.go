package ranker

import (
	"sort"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/process"
)

type CatalogRanker struct {
	catalog *catalog.Catalog
}

func NewCatalogRanker(c *catalog.Catalog) *CatalogRanker {
	return &CatalogRanker{
		catalog: c,
	}
}

// RankCoProcess from newest to oldest based on the introducedBy revision of the coProcess
func (r *CatalogRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	sort.SliceStable(cps, func(i, j int) bool {
		metaI, errI := r.catalog.GetSymbolMetadata(cps[i].Digest)
		metaJ, errJ := r.catalog.GetSymbolMetadata(cps[j].Digest)

		if errI != nil || errJ != nil {
			return false
		}

		return r.catalog.CompareRevisions(metaI.IntroducedBy, metaJ.IntroducedBy) > 0
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

// RankProcess from newest to oldest based on the introducedBy revision of the process
func (r *CatalogRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
	sort.SliceStable(ps, func(i, j int) bool {
		metaI, errI := r.catalog.GetSymbolMetadata(ps[i].Digest)
		metaJ, errJ := r.catalog.GetSymbolMetadata(ps[j].Digest)

		if errI != nil || errJ != nil {
			return false
		}

		return r.catalog.CompareRevisions(metaI.IntroducedBy, metaJ.IntroducedBy) > 0
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
