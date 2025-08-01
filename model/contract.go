package model

import "errors"

type Contract struct {
	Qualifier string `json:"qualifier" yaml:"qualifier"`
	Name      string `json:"name" yaml:"name"`
	Params    map[string]any
}

// TODO: implement attrs?
func (c *Contract) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return c, nil
}
