package hcl

import (
	"encoding/csv"
	"errors"
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

func (c Core) Count(root []string) (map[string]float64, error) {
	sym, err := c.Resolve(NewParserContext(), root)
	if err != nil {
		slog.Info("Unknown symbol.", "error", err, "ref", root)
		return nil, err
	}

	item, ok := sym.(*model.Item)
	if !ok {
		return nil, errors.New("unknown item")
	}

	tree, err := c.Build(item)
	if err != nil {
		return nil, err
	}

	return tree.Count(), nil
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
