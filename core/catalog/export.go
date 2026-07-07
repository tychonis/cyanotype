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
	symbols, err := c.GetSymbols()
	if err != nil {
		return nil, err
	}
	doc.Symbols = symbols
	return json.Marshal(doc)
}
