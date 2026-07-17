package revision

import (
	"errors"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

type Cache struct {
	revisionOrderCache map[model.RevisionID]int
	orderedRevisions   []model.RevisionID
	LatestRevision     model.RevisionID
}

func NewFromIndex(idx map[model.RevisionID]*model.Revision) (*Cache, error) {
	cache := &Cache{
		revisionOrderCache: make(map[model.RevisionID]int),
		orderedRevisions:   make([]model.RevisionID, 0),
	}
	allRevisions := make([]*model.Revision, 0, len(idx))
	for _, rev := range idx {
		allRevisions = append(allRevisions, rev)
	}
	if len(allRevisions) == 0 {
		return cache, nil
	}
	sorted, err := StableTopoRevisions(allRevisions)
	if err != nil {
		return cache, fmt.Errorf("rank revisions: %w", err)
	}
	cache.orderedRevisions = sorted
	cache.LatestRevision = sorted[len(sorted)-1]
	for i, rev := range sorted {
		cache.revisionOrderCache[rev] = i
	}
	return cache, nil
}

func (c *Cache) GetAllRevisions() ([]model.RevisionID, error) {
	return c.orderedRevisions, nil
}

func (c *Cache) GetNewerRevisions(r model.RevisionID) ([]model.RevisionID, error) {
	order, ok := c.revisionOrderCache[r]
	if !ok {
		return nil, errors.New("symbol not found")
	}
	return c.orderedRevisions[order+1:], nil
}

func (c *Cache) CompareRevisions(a, b model.RevisionID) int {
	revAOrder, ok := c.revisionOrderCache[a]
	if !ok {
		return -1
	}
	revBOrder, ok := c.revisionOrderCache[b]
	if !ok {
		return 1
	}
	return revAOrder - revBOrder
}
