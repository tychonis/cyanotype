package ranker

import (
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/model"
)

type CatalogTypeRanker struct {
	PreferedType  string
	catalogRanker *CatalogRanker
}

func NewCatalogTypeRanker(catalog *catalog.Catalog, preferedType string) *CatalogTypeRanker {
	return &CatalogTypeRanker{
		PreferedType:  preferedType,
		catalogRanker: NewCatalogRanker(catalog),
	}
}

func (r *CatalogTypeRanker) RankCoProcess(cps []*model.CoProcess) ([]*model.CoProcess, error) {
	return r.catalogRanker.RankCoProcess(cps)
}

func (r *CatalogTypeRanker) TopCoProcess(cps []*model.CoProcess) (*model.CoProcess, error) {
	return r.catalogRanker.TopCoProcess(cps)
}

func (r *CatalogTypeRanker) RankProcess(ps []*model.Process) ([]*model.Process, error) {
	candidate := make([]*model.Process, 0, len(ps))

	for _, p := range ps {
		if p != nil && p.Get("purpose") == r.PreferedType {
			candidate = append(candidate, p)
		}
	}

	if len(candidate) > 0 {
		return r.catalogRanker.RankProcess(candidate)
	}

	return r.catalogRanker.RankProcess(ps)
}

func (r *CatalogTypeRanker) TopProcess(ps []*model.Process) (*model.Process, error) {
	ranked, err := r.RankProcess(ps)
	if err != nil {
		return nil, err
	}

	if len(ranked) > 0 {
		return ranked[0], nil
	}
	return nil, ErrNotFound
}
