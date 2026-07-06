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
	Name          string `json:"name"`
	Version       string `json:"version"`
	SourceVersion string `json:"source_version"`
	SourceState   string `json:"source_state"`
	Sequence      int    `json:"sequence"`
	UniqueParts   int    `json:"unique_parts"`
}

func GetCatalogMetadata(client *http.Client, endpoint string, tag string) (*CatalogMetadata, error) {
	metaDataEndpoint := fmt.Sprintf("%s/workspace/%s", endpoint, tag)
	resp, err := client.Get(metaDataEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}

	var metadata CatalogMetadata
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (c *Catalog) SaveCatalogMetadata(client *http.Client, endpoint string, tag string) error {
	metaDataEndpoint := fmt.Sprintf("%s/workspace/%s", endpoint, tag)
	symbols, err := c.index.ListSymbols()
	if err != nil {
		return err
	}
	metadata := CatalogMetadata{
		Name:        "placeholder",
		UniqueParts: len(symbols) / 4,
		Sequence:    c.sequence,
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	resp, err := client.Post(metaDataEndpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response")
	}
	return nil
}

func (c *Catalog) Save(endpoint string, token string, tag string) error {
	localIndex, ok := c.index.(*LocalIndex)
	if !ok {
		return errors.New("can only save local index now")
	}

	client := NewClient(token)
	err := c.SaveCatalogMetadata(client, endpoint, tag)
	if err != nil {
		return err
	}

	remote := RemoteIndexFromLocal(localIndex)
	remote.Endpoint = endpoint + "/bom_index/" + tag
	err = remote.Save()
	if err != nil {
		return err
	}

	storage := NewAPIStore(endpoint, client)
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
