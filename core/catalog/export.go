package catalog

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model"
)

type CatalogDocument struct {
	Symbols map[model.Digest]model.ConcreteSymbol `json:"symbols"`
}

func (c *Catalog) Export() ([]byte, error) {
	doc := &CatalogDocument{
		Symbols: make(map[model.Digest]model.ConcreteSymbol),
	}
	symbols, err := c.index.ListSymbols()
	if err != nil {
		return nil, err
	}
	for symDigest := range symbols {
		sym, err := c.Get(symDigest)
		if err != nil {
			return nil, err
		}
		doc.Symbols[symDigest] = sym
	}
	return json.Marshal(doc)
}
