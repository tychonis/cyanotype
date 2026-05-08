package process

import (
	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

func init() {
	processContentTypes["abstract"] = func() ProcessContent { return &Abstract{} }
}

const ABSTRACT = "abstract"

type Abstract struct {
	Name   string           `json:"name" yaml:"name"`
	Input  []*model.BOMLine `json:"input" yaml:"input"`
	Output []*model.BOMLine `json:"output" yaml:"output"`

	Details stable.Map `json:"details" yaml:"details"`
}

func (a Abstract) MarshalJSON() ([]byte, error) {
	type Alias Abstract
	return serializer.JSONWithKey(Alias(a), "type", ABSTRACT)
}

func (a *Abstract) GetName() string {
	return a.Name
}

func (a *Abstract) GetType() string {
	return ABSTRACT
}

func (a *Abstract) GetInput() []*model.BOMLine {
	return a.Input
}

func (a *Abstract) GetOutput() []*model.BOMLine {
	return a.Output
}
