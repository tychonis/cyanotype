package model

import (
	"errors"
)

type ProcessID = Digest

type BOMLine struct {
	Name string  `json:"name" yaml:"name"`
	Item ItemID  `json:"item" yaml:"item"`
	Role string  `json:"role" yaml:"role"`
	Qty  float64 `json:"qty" yaml:"qty"`
}

type ProcessBase struct {
	Qualifier string         `json:"qualifier" yaml:"qualifier"`
	Content   ProcessContent `json:"content" yaml:"content"`

	Digest ProcessID `json:"-" yaml:"-"`
}

type Process struct {
	ProcessBase
}

type CoProcess struct {
	ProcessBase
}

type ProcessContent interface {
	GetName() string
	GetType() string
	GetInput() []*BOMLine
	GetOutput() []*BOMLine
}

func (p *Process) Input() []*BOMLine {
	return p.Content.GetInput()
}

func (p *Process) Output() []*BOMLine {
	return p.Content.GetOutput()
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

func (p *Process) GetQualifier() string {
	return p.Qualifier
}

func (p *Process) GetDigest() string {
	return p.Digest
}

func (cp *CoProcess) Input() []*BOMLine {
	return cp.Content.GetInput()
}

func (cp *CoProcess) Output() []*BOMLine {
	return cp.Content.GetOutput()
}

func (cp *CoProcess) Resolve(path []string) (Symbol, error) {
	if len(path) == 0 {
		return cp, nil
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

func (cp *CoProcess) GetQualifier() string {
	return cp.Qualifier
}

func (cp *CoProcess) GetDigest() string {
	return cp.Digest
}
