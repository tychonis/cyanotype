package hcl

import (
	"strings"

	"github.com/tychonis/cyanotype/model"
)

func (g *BOMGraph) BuildIndex() {
	g.buildPartNumberIdx()
	g.buildQualifierIdx()
	g.buildPathIdx()
}

func (g *BOMGraph) buildPartNumberIdx() {
	for _, item := range g.Items {
		partNumber := item.GetPartNumber()
		if partNumber == "" {
			partNumber = g.generatePartNumber(item)
			item.PartNumber = partNumber
		}
		g.PartNumberIndex[partNumber] = item.GetID()
	}
}

func (g *BOMGraph) buildQualifierIdx() {
	for _, item := range g.Items {
		g.QualifierIndex[item.Qualifier] = item.GetID()
	}
}

func (g *BOMGraph) buildPathIdx() {
	for _, node := range g.Nodes {
		g.PathIndex[node.Path] = node.ID
	}
}

// TODO: In the future, consider support user-defined part numbers with structured revision/variant metadata.
func (g *BOMGraph) generatePartNumber(item model.BOMItem) string {
	id := item.GetID()
	short := strings.Split(id.String(), "-")[0]
	existing, ok := g.PartNumberIndex[short]
	if ok && existing != id {
		const suffixes = "ghijklmnopqrstuvwxyz" // non-hex chars
		for i := 0; i < len(suffixes); i++ {
			candidate := short[:len(short)-1] + string(suffixes[i])
			existing, ok := g.PartNumberIndex[candidate]
			if !ok || existing == id {
				return candidate
			}
		}
	}
	return short
}
