package model

import (
	"errors"
)

type ProcessID = Digest

type BOMLine struct {
	Item ItemID  `json:"item" yaml:"item"`
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

type CoProcess struct {
	Qualifier   string     `json:"qualifier" yaml:"qualifier"`
	Predecessor ProcessID  `json:"predecessor" yaml:"predecessor"`
	CycleTime   float64    `json:"cycle_time" yaml:"cycle_time"`
	Input       []*BOMLine `json:"input" yaml:"input"`
	Output      []*BOMLine `json:"output" yaml:"output"`

	Digest ProcessID `json:"-" yaml:"-"`
}

type ProcessContent struct {
	Name            string             `json:"name" yaml:"name"`
	Transformations []TransformationID `json:"transformations" yaml:"transformations"`
}

func (p *Process) Resolve(path []string) (Symbol, error) {
	if len(path) == 0 {
		return p, nil
	}
	if len(path) < 2 {
		return nil, errors.New("insufficient path length")
	}
	switch path[0] {
	case "input":
		return nil, nil
	case "output":
		return nil, nil
	default:
		return nil, errors.New("illformed token")
	}
}
