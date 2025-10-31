package catalog

import (
	"encoding/json"
	"errors"

	"github.com/tychonis/cyanotype/model"
)

type ItemInfo struct {
	Name string `json:"name"`
}

type CatalogDocument struct {
	Items map[model.Digest]*ItemInfo `json:"items"`
}

func (c *Catalog) Export() ([]byte, error) {
	doc := &CatalogDocument{
		Items: make(map[model.Digest]*ItemInfo),
	}
	symbols, err := c.index.ListSymbols()
	if err != nil {
		return nil, err
	}
	for symDigest, symType := range symbols {
		if symType == "item" {
			sym, err := c.Get(symDigest)
			if err != nil {
				return nil, err
			}
			item, ok := sym.(*model.Item)
			if !ok {
				return nil, errors.New("unexpected type")
			}
			info := &ItemInfo{
				Name: item.Content.Name,
			}
			doc.Items[symDigest] = info
		}
	}
	return json.Marshal(doc)
}
