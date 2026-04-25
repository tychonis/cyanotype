package process

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

type Abstract struct {
	Name   string           `json:"name" yaml:"name"`
	Input  []*model.BOMLine `json:"input" yaml:"input"`
	Output []*model.BOMLine `json:"output" yaml:"output"`

	Details stable.Map `json:"details" yaml:"details"`
}

func withType(t string, v any) ([]byte, error) {
	m := map[string]any{
		"type": t,
	}
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &m)
	return json.Marshal(m)
}

func (a Abstract) MarshalJSON() ([]byte, error) {
	return withType("abstract", a)
}

func (a *Abstract) GetName() string {
	return a.Name
}

func (a *Abstract) GetType() string {
	return "abstract"
}

func (a *Abstract) GetInput() []*model.BOMLine {
	return a.Input
}

func (a *Abstract) GetOutput() []*model.BOMLine {
	return a.Output
}
