package model

import (
	"errors"

	"github.com/tychonis/cyanotype/model"
)

type ProcessID = Digest

type BOMLine struct {
	ID   ItemID  `json:"id" yaml:"id"`
	Qty  float64 `json:"qty" yaml:"qty"`
	Role string  `json:"role" yaml:"role"`
}

type Process struct {
	Qualifier   string     `json:"qualifier" yaml:"qualifier"`
	Predecessor ProcessID  `json:"predecessor" yaml:"predecessor"`
	CycleTime   float64    `json:"cycle_time" yaml:"cycle_time"`
	Input       []*BOMLine `json:"input" yaml:"input"`
	Output      []*BOMLine `json:"output" yaml:"output"`

	Digest ProcessID `json:"-" yaml:"-"`
}

type ProcessContent struct {
	Name            string   `json:"name" yaml:"name"`
	Transformations []string `json:"transformations" yaml:"transformations"`
}

// TODO: implement attrs?
func (p *Process) Resolve(path []string) (model.Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return p, nil
}
