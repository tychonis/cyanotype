package hcl

import (
	"strings"

	"github.com/tychonis/cyanotype/model"
)

func (g *BOMGraph) BuildCatalog() {
	g.buildNameIdx()
	g.buildPartNumberIdx()
	g.buildCatalogDetails()
}

func (g *BOMGraph) buildNameIdx() {
	for _, item := range g.Items {
		g.Catalog.NameIdx[item.GetName()] = item.GetID()
	}
}

func (g *BOMGraph) buildPartNumberIdx() {
	for _, item := range g.Items {
		partNumber := item.GetPartNumber()
		if partNumber == "" {
			partNumber = g.generatePartNumber(item)
			item.PartNumber = partNumber
		}
		g.Catalog.PartNumberIdx[partNumber] = item.GetID()
	}
}

func (g *BOMGraph) buildItems(items Items) {
	for _, item := range items {
		g.Catalog.Catalog[item.GetID()] = item
	}
}

func (g *BOMGraph) buildCatalogDetails() {
	g.buildItems(g.Items)
	for _, items := range g.Variants {
		g.buildItems(items)
	}
}

// TODO: In the future, consider support user-defined part numbers with structured revision/variant metadata.
func (g *BOMGraph) generatePartNumber(item model.BOMItem) string {
	id := item.GetID()
	short := strings.Split(id.String(), "-")[0]
	existing, ok := g.Catalog.PartNumberIdx[short]
	if ok && existing != id {
		const suffixes = "ghijklmnopqrstuvwxyz" // non-hex chars
		for i := 0; i < len(suffixes); i++ {
			candidate := short[:len(short)-1] + string(suffixes[i])
			existing, ok := g.Catalog.PartNumberIdx[candidate]
			if !ok || existing == id {
				return candidate
			}
		}
	}
	return short
}
