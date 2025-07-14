package states

import (
	"errors"
	"log/slog"
	"maps"

	"github.com/google/uuid"
)

type Catalog struct {
	ID            uuid.UUID            `json:"id"`
	Version       string               `json:"version"`
	NameIdx       map[string]uuid.UUID `json:"name_index"`
	PartNumberIdx map[string]uuid.UUID `json:"part_number_index"`
}

func NewCatalog() *Catalog {
	return &Catalog{
		ID:            uuid.New(),
		Version:       "alpha-0",
		NameIdx:       make(map[string]uuid.UUID),
		PartNumberIdx: make(map[string]uuid.UUID),
	}
}

func (c *Catalog) MergeCatalog(c2 *Catalog) error {
	if c2.Version != c.Version {
		slog.Warn("Merging incompatible catalogs.",
			"src_version", c2.Version, "dst_version", c.Version)
		return errors.New("merging incompatible catalogs")
	}
	// TODO: handle key conflict.
	maps.Copy(c.NameIdx, c2.NameIdx)
	maps.Copy(c.PartNumberIdx, c2.PartNumberIdx)
	return nil
}
