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
	for name, item := range g.Items {
		g.Catalog.NameIdx[name] = item.GetID()
	}
}

func (g *BOMGraph) buildPartNumberIdx() {
	for _, item := range g.Items {
		partNumber := item.GetPartNumber()
		if partNumber == "" {
			part, ok := item.(*model.Item)
			if ok {
				partNumber = g.generatePartNumber(item)
				part.PartNumber = partNumber
			}
		}
		g.Catalog.PartNumberIdx[partNumber] = item.GetID()
	}
}

func (g *BOMGraph) buildCatalogDetails() {
	for _, item := range g.Items {
		details := item.GetDetails()
		details["name"] = item.GetName()
		details["part_number"] = item.GetPartNumber()
		g.Catalog.Catalog[item.GetID()] = details
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
