package operation

import (
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

func AlignOperation(points []model.ConcreteSymbol) *model.Operation {
	op := &model.Operation{
		Category: "align",
		Params:   make(map[string]interface{}),
	}
	qualifiers := make([]model.Qualifier, 0, len(points))
	for _, point := range points {
		qualifiers = append(qualifiers, point.GetQualifier())
	}
	op.Params["points"] = qualifiers
	return op
}

func AlignFromOperation(op *model.Operation) ([]model.Qualifier, error) {
	points, ok := op.Params["points"].([]model.Qualifier)
	if !ok {
		return nil, fmt.Errorf("points not found or invalid")
	}
	return points, nil
}
