package operation

import (
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

func AlignOperation(points []model.Qualifier) *model.Operation {
	op := &model.Operation{
		Category: "align",
		Params:   make(map[string]interface{}),
	}
	qualifiers := make([]model.Qualifier, 0, len(points))
	for _, point := range points {
		qualifiers = append(qualifiers, point)
	}
	op.Params["points"] = qualifiers
	return op
}

func AlignFromOperation(op *model.Operation) ([]model.Qualifier, error) {
	points, ok := op.Params["points"].([]any)
	if !ok {
		return nil, fmt.Errorf("points not found or invalid")
	}
	pointsSlice := make([]model.Qualifier, 0, len(points))
	for _, point := range points {
		q, ok := point.(model.Qualifier)
		if !ok {
			return nil, fmt.Errorf("invalid point in points")
		}
		pointsSlice = append(pointsSlice, q)
	}
	return pointsSlice, nil
}
