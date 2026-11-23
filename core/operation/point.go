package operation

import (
	"encoding/json"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

func PointOperation(p model.Placement) *model.Operation {
	op := &model.Operation{
		Category: "point",
		Params:   make(map[string]interface{}),
	}
	op.Params["placement"] = p
	return op
}

func PointFromOperation(op *model.Operation) (model.Placement, error) {
	if op.Category != "point" {
		return model.IdentityPlacement, fmt.Errorf("operation is not a point")
	}
	return extractPlacement(op)
}

func extractPlacement(op *model.Operation) (model.Placement, error) {
	value, ok := op.Params["placement"]
	if !ok {
		return model.IdentityPlacement, fmt.Errorf("point placement not found")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return model.IdentityPlacement, err
	}

	var placement model.Placement
	if err := json.Unmarshal(data, &placement); err != nil {
		return model.IdentityPlacement, err
	}

	return placement, nil
}
