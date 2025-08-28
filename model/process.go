package model

import (
	"errors"

	"github.com/google/uuid"
)

type ProcessID = uuid.UUID

type Process struct {
	ID             ProcessID `json:"id" yaml:"id"`
	Qualifier      string    `json:"qualifier" yaml:"qualifier"`
	Name           string    `json:"name" yaml:"name"`
	Input          []*Component
	Output         []*Component
	Transformation func([]*Contract) *[]Contract
}

// TODO: implement attrs?
func (p *Process) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return p, nil
}
