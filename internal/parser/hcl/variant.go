package hcl

import (
	"log/slog"

	"github.com/tychonis/cyanotype/model"
)

func (g *BOMGraph) InstantiateVariant(variant string) *BOMGraph {
	ng := NewBOMGraph()
	ng.MergeGraph(g)
	for name, item := range ng.Items {
		asm, ok := item.(*model.Item)
		if !ok {
			continue
		}
		if asm.Source != "virtual" {
			continue
		}
		imp, ok := ng.Variants[variant][asm.ID]
		if !ok {
			slog.Warn("Unimplemented variant.", "part", asm.Name, "variant", variant)
			continue
		}
		ng.Items[name] = imp
	}
	ng.Build()
	return ng
}
