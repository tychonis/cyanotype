package instantiator

type Component struct {
	Name string
	Ref  []string
	Qty  float64
}

func (i *Instantiator) Count(root string) (map[string]float64, error) {
	tree, err := i.TreeFromQualifier(root)
	if err != nil {
		return nil, err
	}

	return tree.Count(), nil
}
