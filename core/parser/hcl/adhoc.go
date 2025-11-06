package hcl

import (
	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/internal/ranker"
	"github.com/tychonis/cyanotype/internal/symbols"
)

func NewCoreFromAPI(endpoint string, tag string) *Core {
	return &Core{
		Symbols: symbols.NewSymbolTable(),
		Catalog: catalog.NewRemoteCatalog(endpoint, tag),

		Ranker: &ranker.NaiveRanker{},
	}
}

// Adhoc function supporting bomhub.
func (c *Core) SaveCatalog(endpoint string, tag string) error {
	return c.Catalog.Save(endpoint, tag)
}
