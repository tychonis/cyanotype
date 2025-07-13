package hcl

import (
	"log/slog"
)

func (g *BOMGraph) InstantiateVariant(variant string) *BOMGraph {
	ng := NewBOMGraph()
	ng.MergeGraph(g)
	for name, item := range ng.Items {
		if item.Source != "virtual" {
			continue
		}
		imp, ok := ng.Variants[variant][item.ID]
		if !ok {
			slog.Warn("Unimplemented variant.", "part", item.Name, "variant", variant)
			continue
		}
		ng.Items[name] = imp
	}
	ng.Build()
	return ng
}
