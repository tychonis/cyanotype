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
	var revision model.RevisionID
	if c.latestRevision == nil {
		revision = ""
	} else {
		revision = c.latestRevision.Digest
	}
	doc := &CatalogDocument{
		Revision: revision,
		Symbols:  make(map[model.Digest]model.ConcreteSymbol),
	}
	symbols, err := c.GetSymbols()
	if err != nil {
		return nil, err
	}
	doc.Symbols = symbols
	return json.Marshal(doc)
}
