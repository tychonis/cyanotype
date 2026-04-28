package process

import (
	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

func init() {
	processContentTypes["drawing"] = func() ProcessContent { return &Drawing{} }
}

type Component struct {
	Name        string       `json:"name" yaml:"name"`
	CoItem      model.ItemID `json:"coitem" yaml:"coitem"`
	Rotation    [4]float64   `json:"rotation" yaml:"rotation"`
	Translation [3]float64   `json:"translation" yaml:"translation"`
}

type Drawing struct {
	Name       string           `json:"name" yaml:"name"`
	Components []*Component     `json:"components" yaml:"components"`
	Output     []*model.BOMLine `json:"output" yaml:"output"`

	Details stable.Map `json:"details" yaml:"details"`
}

func (d Drawing) MarshalJSON() ([]byte, error) {
	type Alias Drawing
	return serializer.JSONWithKey(Alias(d), "type", "drawing")
}

func (d *Drawing) GetName() string {
	return d.Name
}

func (d *Drawing) GetType() string {
	return "drawing"
}

func (d *Drawing) GetInput() []*model.BOMLine {
	ret := make([]*model.BOMLine, 0, len(d.Components))
	for _, component := range d.Components {
		ret = append(ret, &model.BOMLine{
			Name: component.Name,
			Item: component.CoItem,
			Qty:  1,
		})
	}
	return ret
}

func (d *Drawing) GetOutput() []*model.BOMLine {
	return d.Output
}
