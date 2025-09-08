package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tychonis/cyanotype/model"
)

type ContractID = uuid.UUID

type Contract struct {
	ID        ContractID `json:"id" yaml:"id"`
	Qualifier string     `json:"qualifier" yaml:"qualifier"`
	Name      string     `json:"name" yaml:"name"`
	Params    map[string]any
}

// TODO: implement attrs?
func (c *Contract) Resolve(path []string) (model.Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return c, nil
}
