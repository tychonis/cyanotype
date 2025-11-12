package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tychonis/cyanotype/internal/serializer"
)

type CatalogMetadata struct {
	Name        string `json:"name"`
	UniqueParts int    `json:"unique_parts"`
}

func (c *Catalog) SaveMetadata(endpoint string, tag string) error {
	metaDataEndpoint := fmt.Sprintf("%s/workspace/%s", endpoint, tag)
	symbols, err := c.index.ListSymbols()
	if err != nil {
		return err
	}
	metadata := CatalogMetadata{
		Name:        "placeholder",
		UniqueParts: len(symbols),
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	resp, err := http.Post(metaDataEndpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response")
	}
	return nil
}

func (c *Catalog) Save(endpoint string, tag string) error {
	localIndex, ok := c.index.(*LocalIndex)
	if !ok {
		return errors.New("can only save local index now")
	}

	err := c.SaveMetadata(endpoint, tag)
	if err != nil {
		return err
	}

	remote := RemoteIndexFromLocal(localIndex)
	remote.Endpoint = endpoint + "/bom_index/" + tag
	err = remote.Save()
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
