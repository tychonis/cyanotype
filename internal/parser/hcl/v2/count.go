package hcl

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"

	"github.com/tychonis/cyanotype/model/v2"
)

type Component struct {
	Name string
	Ref  []string
	Qty  float64
}

func (c *Core) getComponents(item *model.Item) []*Component {
	return nil
}

func (c *Core) countParts(ref []string, multiplier float64, counter map[string]float64) {
	slog.Debug("Counting...", "name", ref, "multiplier", multiplier)
	sym, err := c.Resolve(NewParserContext(), ref)
	if err != nil {
		slog.Info("Unknown symbol.", "error", err, "ref", ref)
		return
	}

	item, ok := sym.(*model.Item)
	if !ok {
		slog.Info("Unknown item.", "error", err, "ref", ref)
		return
	}

	components := c.getComponents(item)

	if len(components) == 0 {
		counter[item.Qualifier] += multiplier
		return
	}

	// also count assembly?
	counter[item.Qualifier] += multiplier

	for _, comp := range components {
		c.countParts(comp.Ref, comp.Qty*multiplier, counter)
	}
}

func (c Core) Count(root []string) map[string]float64 {
	counter := make(map[string]float64)
	c.countParts(root, 1, counter)

	return counter
}

func getHeader() []string {
	return []string{"Part ID", "Part Number", "Name", "Quantity"}
}

func (c *Core) CounterToCSV(counter map[string]float64) {
	writer := csv.NewWriter(os.Stdout)
	writer.Write(getHeader())
	for name, qty := range counter {
		line := []string{name,
			fmt.Sprintf("%.2f", qty),
		}
		writer.Write(line)
	}
	writer.Flush()
}
