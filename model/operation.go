package model

import (
	"github.com/tychonis/cyanotype/internal/stable"
)

type OperationID = Digest

type Operation struct {
	SymbolBase

	Category string     `json:"category" yaml:"category"`
	Name     string     `json:"name" yaml:"name"`
	Params   stable.Map `json:"params" yaml:"params"`
}

func (c *Operation) GetType() string {
	return "operation"
}
