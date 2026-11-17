package model

import (
	"github.com/tychonis/cyanotype/internal/stable"
)

type ContractID = Digest

type Contract struct {
	SymbolBase

	Category string     `json:"category" yaml:"category"`
	Name     string     `json:"name" yaml:"name"`
	Params   stable.Map `json:"params" yaml:"params"`
}

func (c *Contract) Fulfill(c2 *Contract) bool {
	return false
}

func (c *Contract) GetType() string {
	return "contract"
}
