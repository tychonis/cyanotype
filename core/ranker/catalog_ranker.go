package ranker

import (
	"sort"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/model"
)

type CatalogRanker struct {
	catalog *catalog.Catalog
}

func NewCatalogRanker(c *catalog.Catalog) *CatalogRanker {
	return &CatalogRanker{
		catalog: c,
	}
}

// RankCoProcess from newest to oldest based on the LastCommited revision of the coProcess
func (r *CatalogRanker) RankCoProcess(cps []*model.CoProcess) ([]*model.CoProcess, error) {
	sort.SliceStable(cps, func(i, j int) bool {
		metaI, errI := r.catalog.GetSymbolMetadata(cps[i].Digest)
		metaJ, errJ := r.catalog.GetSymbolMetadata(cps[j].Digest)

		if errI != nil || errJ != nil {
			return false
		}

		return r.catalog.CompareRevisions(metaI.LastCommitted(), metaJ.LastCommitted()) > 0
	})
	return cps, nil
}

func (r *CatalogRanker) TopCoProcess(cps []*model.CoProcess) (*model.CoProcess, error) {
	if len(cps) > 0 {
		cps, err := r.RankCoProcess(cps)
		if err != nil {
			return nil, err
		}
		return cps[0], nil
	}
	return nil, ErrNotFound
}

// RankProcess from newest to oldest based on the LastCommited revision of the process
func (r *CatalogRanker) RankProcess(ps []*model.Process) ([]*model.Process, error) {
	sort.SliceStable(ps, func(i, j int) bool {
		metaI, errI := r.catalog.GetSymbolMetadata(ps[i].Digest)
		metaJ, errJ := r.catalog.GetSymbolMetadata(ps[j].Digest)

		if errI != nil || errJ != nil {
			return false
		}

		return r.catalog.CompareRevisions(metaI.LastCommitted(), metaJ.LastCommitted()) > 0
	})
	return ps, nil
}

func (r *CatalogRanker) TopProcess(ps []*model.Process) (*model.Process, error) {
	if len(ps) > 0 {
		ps, err := r.RankProcess(ps)
		if err != nil {
			return nil, err
		}
		return ps[0], nil
	}
	return nil, ErrNotFound
}
