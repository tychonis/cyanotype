package process

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

var processContentTypes = make(map[string]func() ProcessContent)

type ProcessID = model.Digest

type ProcessBase struct {
	Type      string         `json:"type" yaml:"type"`
	Qualifier string         `json:"qualifier" yaml:"qualifier"`
	Content   ProcessContent `json:"content" yaml:"content"`

	Digest ProcessID `json:"-" yaml:"-"`
}

func (pb *ProcessBase) UnmarshalJSON(data []byte) error {
	type Alias ProcessBase

	var aux struct {
		*Alias
		Content json.RawMessage `json:"content"`
	}

	aux.Alias = (*Alias)(pb)

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var probe struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(aux.Content, &probe); err != nil {
		return err
	}

	ctor, ok := processContentTypes[probe.Type]
	if !ok {
		return fmt.Errorf("unknown content type %q", probe.Type)
	}

	v := ctor()
	if err := json.Unmarshal(aux.Content, v); err != nil {
		return err
	}

	pb.Content = v
	return nil
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
	GetInput() []*model.BOMLine
	GetOutput() []*model.BOMLine
}

func (p *Process) Input() []*model.BOMLine {
	return p.Content.GetInput()
}

func (p *Process) Output() []*model.BOMLine {
	return p.Content.GetOutput()
}

func (p *Process) Resolve(path []string) (model.Symbol, error) {
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

func (p *Process) GetType() string {
	return p.Type
}

func (cp *CoProcess) Input() []*model.BOMLine {
	return cp.Content.GetInput()
}

func (cp *CoProcess) Output() []*model.BOMLine {
	return cp.Content.GetOutput()
}

func (cp *CoProcess) Resolve(path []string) (model.Symbol, error) {
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

func (cp *CoProcess) GetType() string {
	return cp.Type
}
