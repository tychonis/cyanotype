package model

import (
	"errors"

	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

type ContractID = Digest

type Contract struct {
	Qualifier string     `json:"qualifier" yaml:"qualifier"`
	Name      string     `json:"name" yaml:"name"`
	Params    stable.Map `json:"params" yaml:"params"`

	Digest ContractID `json:"-" yaml:"-"`
}

// TODO: implement attrs?
func (c *Contract) Resolve(path []string) (model.Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return c, nil
}
