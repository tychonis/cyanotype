package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tychonis/cyanotype/model"
)

type CatalogMetadata struct {
	Name           string          `json:"name"`
	LatestRevision *model.Revision `json:"latest_revision"`
	UniqueParts    int             `json:"unique_parts"`
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
	metadata := CatalogMetadata{
		Name:           "placeholder",
		LatestRevision: c.latestRevision,
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

// func (c *Catalog) Save(endpoint string, token string, tag string) error {
// 	localIndex, ok := c.index.(*LocalIndex)
// 	if !ok {
// 		return errors.New("can only save local index now")
// 	}

// 	client := NewHTTPClient(token)

// 	remote := RemoteIndexFromLocal(localIndex)
// 	remote.Endpoint = endpoint + "/bom_index/" + tag
// 	err := remote.Save()
// 	if err != nil {
// 		return err
// 	}

// 	remoteCatalog := NewRemoteCatalog(endpoint, token, tag)
// 	symbols, err := c.GetSymbols()
// 	if err != nil {
// 		return err
// 	}
// 	for symDigest := range symbols {
// 		sym, err := c.Get(symDigest)
// 		if err != nil {
// 			return err
// 		}
// 		data, err := serializer.Serialize(sym)
// 		if err != nil {
// 			return err
// 		}
// 		remoteSym, _ := remoteCatalog.Get(symDigest)
// 		// TODO: This is a hacky way to check if the symbol exists in the remote catalog.
// 		// use a proper error type to check for this instead of relying on nil.
// 		if remoteSym != nil {
// 			continue
// 		}
// 		err = remoteCatalog.storage.Save(symDigest, data)
// 		if err != nil {
// 			return err
// 		}
// 		metadata, err := c.GetMetadata(symDigest)
// 		if err != nil {
// 			return err
// 		}
// 		metadataBytes, err := json.Marshal(metadata)
// 		if err != nil {
// 			return err
// 		}
// 		err = remoteCatalog.storage.SaveMetadata(symDigest, metadataBytes)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return c.SaveCatalogMetadata(client, endpoint, tag)
// }
