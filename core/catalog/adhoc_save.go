package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tychonis/cyanotype/internal/serializer"
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

func (c *Catalog) GetNewerRevisions(base *model.Revision) ([]*model.Revision, error) {
	var newRevisions []model.RevisionID
	var err error
	if base == nil {
		newRevisions, err = c.index.GetAllRevisions()
	} else {
		newRevisions, err = c.index.GetNewerRevisions(base.Digest)
	}
	if err != nil {
		return nil, err
	}
	ret := make([]*model.Revision, 0, len(newRevisions))
	for _, revID := range newRevisions {
		body, err := c.storage.Load(revID)
		if err != nil {
			return nil, err
		}
		rev, err := serializer.Deserialize[*model.Revision](body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rev)
	}
	return ret, nil
}

func (c *Catalog) Pull(other *Catalog) error {
	newRevisions, err := other.GetNewerRevisions(c.latestRevision)
	if err != nil {
		return err
	}
	if len(newRevisions) == 0 {
		return errors.New("other catalog has no newer revisions")
	}
	for _, rev := range newRevisions {
		c.index.IndexRevision(rev)
		body, err := serializer.Serialize(rev)
		if err != nil {
			return err
		}
		err = c.storage.Save(rev.Digest, body)
		if err != nil {
			return err
		}
	}
	allSymbols, err := other.index.GetAllSymbols()
	if err != nil {
		return err
	}
	for _, symDigest := range allSymbols {
		sym, err := other.Get(symDigest)
		if err != nil {
			return err
		}
		metadata, err := other.GetMetadata(symDigest)
		if err != nil {
			return err
		}
		if c.latestRevision == nil || other.index.CompareRevisions(metadata.IntroducedBy, c.latestRevision.Digest) > 0 {
			revData, err := other.storage.Load(metadata.IntroducedBy)
			if err != nil {
				return err
			}
			rev, err := serializer.Deserialize[*model.Revision](revData)
			if err != nil {
				return err
			}
			c.Add(rev, sym)
		}
	}
	// TODO: handle save logic elsewhere, maybe in the index itself.
	switch c.index.(type) {
	case *RemoteIndex:
		remoteIdx, _ := c.index.(*RemoteIndex)
		remoteIdx.Save()
	}
	return nil
}

func (c *Catalog) Push(other *Catalog) error {
	return other.Pull(c)
}
