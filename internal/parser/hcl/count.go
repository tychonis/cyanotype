package hcl

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
)

func (g *BOMGraph) countParts(name string, multiplier float64, counter map[string]float64) {
	item, ok := g.Items[name]
	if !ok {
		slog.Info("Unknown items.", "item", name)
		return
	}

	if len(item.GetComponents()) == 0 {
		counter[item.GetName()] += multiplier
		return
	}

	// also count assembly?
	counter[item.GetName()] += multiplier

	for _, comp := range item.GetComponents() {
		g.countParts(comp.Item.GetName(), comp.Qty*multiplier, counter)
	}
}

func (g *BOMGraph) Count(root string) map[string]float64 {
	counter := make(map[string]float64)
	g.countParts(root, 1, counter)

	return counter
}

func getHeader() []string {
	return []string{"Part ID", "Part Number", "Name", "Quantity"}
}

func (g *BOMGraph) CounterToCSV(counter map[string]float64) {
	writer := csv.NewWriter(os.Stdout)
	writer.Write(getHeader())
	for name, qty := range counter {
		line := []string{name,
			g.Items[name].GetID().String(),
			g.Items[name].GetPartNumber(),
			fmt.Sprintf("%.2f", qty),
		}
		writer.Write(line)
	}
	writer.Flush()
}
