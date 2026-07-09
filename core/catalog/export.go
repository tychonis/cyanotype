package catalog

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model"
)

type CatalogDocument struct {
	Revision model.RevisionID                      `json:"revision"`
	Symbols  map[model.Digest]model.ConcreteSymbol `json:"symbols"`
}

func (c *Catalog) Export() ([]byte, error) {
	doc := &CatalogDocument{
		Revision: c.latestRevision.Digest,
		Symbols:  make(map[model.Digest]model.ConcreteSymbol),
	}
	symbols, err := c.GetSymbols()
	if err != nil {
		return nil, err
	}
	doc.Symbols = symbols
	return json.Marshal(doc)
}
