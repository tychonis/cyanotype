package catalog

import "github.com/tychonis/cyanotype/model"

type CatalogStatistics struct {
	TotalItems   int `json:"total_items"`
	TotalCoItems int `json:"total_coitems"`
}

func (c *Catalog) GetStatistics() (*CatalogStatistics, error) {
	allSymbols, err := c.index.GetAllSymbols()
	if err != nil {
		return nil, err
	}
	stats := &CatalogStatistics{}
	for _, symDigest := range allSymbols {
		sym, err := c.Get(symDigest)
		if err != nil {
			return nil, err
		}
		switch sym.(type) {
		case *model.Item:
			stats.TotalItems++
		case *model.CoItem:
			stats.TotalCoItems++
		}
	}
	return stats, nil
}
