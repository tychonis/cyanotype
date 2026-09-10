package instantiator

import (
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
