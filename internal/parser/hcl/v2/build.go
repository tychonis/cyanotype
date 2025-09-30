package hcl

import (
	"errors"

	"github.com/tychonis/cyanotype/model/v2"
)

func getImplicitProcessQualifier(item *model.Item) string {
	return IMPLICIT + item.Qualifier + ".process"
}

func getImplicitCoProcessQualifier(item *model.Item) string {
	return IMPLICIT + item.Qualifier + ".coprocess"
}

func getImplicitCoItemQualifier(item *model.Item) string {
	return IMPLICIT + item.Qualifier + ".coitem"
}

func (c *Core) findImplicitProcess(item *model.Item) (*model.Process, error) {
	q := getImplicitProcessQualifier(item)
	sym, err := c.Catalog.Find(q)
	if err != nil {
		return nil, err
	}
	p, ok := sym.(*model.Process)
	if !ok {
		return nil, errors.New("incorrect type for implicit process")
	}
	return p, nil
}

func (c *Core) Build(item *model.Item) error {
	return nil
}
