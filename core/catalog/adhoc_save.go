package catalog

import (
	"errors"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model"
)

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
	err = c.updateLatestRevision()
	if err != nil {
		return err
	}
	// TODO: handle save logic elsewhere, maybe in the index itself.
	switch c.index.(type) {
	case *RemoteIndex:
		remoteIdx, _ := c.index.(*RemoteIndex)
		err = remoteIdx.Save()
		if err != nil {
			return err
		}
		err = remoteIdx.SaveCatalogMetadata()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Catalog) Push(other *Catalog) error {
	return other.Pull(c)
}

func (c *Catalog) updateLatestRevision() error {
	latestRev, err := c.index.GetLatestRevision()
	if err != nil {
		return err
	}
	if latestRev == nil {
		return nil
	}
	fullRevData, err := c.storage.Load(latestRev.Digest)
	if err != nil {
		return err
	}
	fullRev, err := serializer.Deserialize[*model.Revision](fullRevData)
	if err != nil {
		return err
	}
	c.latestRevision = fullRev
	return nil
}
