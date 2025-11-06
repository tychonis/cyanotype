package catalog

import (
	"errors"

	"github.com/tychonis/cyanotype/internal/serializer"
)

func (c *Catalog) Save(endpoint string, tag string) error {
	localIndex, ok := c.index.(*LocalIndex)
	if !ok {
		return errors.New("can only save local index now")
	}
	remote := RemoteIndexFromLocal(localIndex)
	remote.Endpoint = endpoint + "/bom_index/" + tag
	err := remote.Save()
	if err != nil {
		return err
	}

	storage := NewAPIStore(endpoint + "/obj")
	symbols, err := c.index.ListSymbols()
	if err != nil {
		return err
	}
	for symDigest := range symbols {
		sym, err := c.Get(symDigest)
		if err != nil {
			return err
		}
		data, err := serializer.Serialize(sym)
		if err != nil {
			return err
		}
		err = storage.Save(symDigest, data)
		if err != nil {
			return err
		}
	}
	return nil
}
