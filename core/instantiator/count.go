package instantiator

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/tychonis/cyanotype/core/catalog"
)

type Component struct {
	Name string
	Ref  []string
	Qty  float64
}

func (i *Instantiator) Count(cat *catalog.Catalog, root string) (map[string]float64, error) {
	tree, err := i.TreeFromQualifier(cat, root)
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
