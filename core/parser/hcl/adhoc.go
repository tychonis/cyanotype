package hcl

import (
	"github.com/tychonis/cyanotype/core/ranker"

	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/internal/symbols"
)

func NewCoreFromAPI(endpoint string, token string, tag string) *Core {
	return &Core{
		Symbols: symbols.NewSymbolTable(),
		Catalog: catalog.NewRemoteCatalog(endpoint, token, tag),

		Ranker: &ranker.NaiveRanker{},
	}
}

// Adhoc function supporting bomhub.
func (c *Core) SaveCatalog(endpoint string, token string, tag string) error {
	return c.Catalog.Save(endpoint, token, tag)
}
