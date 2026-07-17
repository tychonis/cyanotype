package instantiator

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/model"
)

type Component struct {
	Name string
	Ref  []string
	Qty  float64
}

func (i *Instantiator) Count(cat *catalog.Catalog, root string) (map[string]float64, error) {
	sym, err := cat.FindCurrent(root)
	if err != nil {
		slog.Info("Unknown symbol.", "error", err, "ref", root)
		return nil, err
	}

	item, ok := sym.(*model.Item)
	if !ok {
		return nil, errors.New("unknown item")
	}

	tree, err := i.InstantiateTreeFromItem(cat, "root", item)
	if err != nil {
		return nil, err
	}

	return tree.Count(), nil
}

func getHeader() []string {
	return []string{"Part ID", "Part Number", "Name", "Quantity"}
}

func (i *Instantiator) CounterToCSV(counter map[string]float64) {
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
