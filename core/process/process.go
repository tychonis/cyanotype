package process

import (
	"encoding/json"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

var processContentTypes = make(map[string]func() ProcessContent)

type ProcessID = model.Digest

type ProcessBase struct {
	model.SymbolBase

	Content ProcessContent `json:"content" yaml:"content"`
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
	Operations []*model.Operation `json:"operations" yaml:"operations"`
}

type CoProcess struct {
	ProcessBase
	Operations []*model.Operation `json:"operations" yaml:"operations"`
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

func (p *Process) GetType() string {
	return "process"
}

func (cp *CoProcess) Input() []*model.BOMLine {
	return cp.Content.GetInput()
}

func (cp *CoProcess) Output() []*model.BOMLine {
	return cp.Content.GetOutput()
}

func (cp *CoProcess) GetType() string {
	return "coprocess"
}
