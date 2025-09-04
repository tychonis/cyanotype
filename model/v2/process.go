package model

import (
	"errors"

	"github.com/google/uuid"
)

type ProcessID = uuid.UUID

type BOMLine struct {
	ID   ItemID  `json:"id" yaml:"id"`
	Qty  float64 `json:"qty" yaml:"qty"`
	Role string  `json:"role" yaml:"role"`
}

type Process struct {
	ID          ProcessID `json:"id" yaml:"id"`
	Qualifier   string    `json:"qualifier" yaml:"qualifier"`
	Predecessor ProcessID
	Input       []*BOMLine
	Output      []*BOMLine
}

type ProcessContent struct {
	Name           string `json:"name" yaml:"name"`
	Transformation func([]*Contract) *[]Contract
}

// TODO: implement attrs?
func (p *Process) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return p, nil
}
