package ranker

import (
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/process"
)

type CatalogTypeRanker struct {
	PreferedType  string
	catalogRanker *CatalogRanker
}

func NewCatalogTypeRanker(preferedType string, catalog *catalog.Catalog) *CatalogTypeRanker {
	return &CatalogTypeRanker{
		PreferedType:  preferedType,
		catalogRanker: NewCatalogRanker(catalog),
	}
}

func (r *CatalogTypeRanker) RankCoProcess(cps []*process.CoProcess) ([]*process.CoProcess, error) {
	return r.catalogRanker.RankCoProcess(cps)
}

func (r *CatalogTypeRanker) TopCoProcess(cps []*process.CoProcess) (*process.CoProcess, error) {
	return r.catalogRanker.TopCoProcess(cps)
}

func (r *CatalogTypeRanker) RankProcess(ps []*process.Process) ([]*process.Process, error) {
	candidate := make([]*process.Process, 0, len(ps))

	for _, p := range ps {
		if p != nil && p.Content != nil && p.Content.GetType() == r.PreferedType {
			candidate = append(candidate, p)
		}
	}

	if len(candidate) > 0 {
		return r.catalogRanker.RankProcess(candidate)
	}

	return r.catalogRanker.RankProcess(ps)
}

func (r *CatalogTypeRanker) TopProcess(ps []*process.Process) (*process.Process, error) {
	ranked, err := r.RankProcess(ps)
	if err != nil {
		return nil, err
	}

	if len(ranked) > 0 {
		return ranked[0], nil
	}
	return nil, ErrNotFound
}
