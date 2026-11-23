package operation

import (
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

func PlacementOperation(child string, p model.Placement) *model.Operation {
	op := &model.Operation{
		Category: "placement",
		Params:   make(map[string]interface{}),
	}
	op.Params["placement"] = p
	op.Params["child"] = child
	return op
}

func PlacementFromOperation(op *model.Operation) (string, model.Placement, error) {
	if op.Category != "placement" {
		return "", model.IdentityPlacement, fmt.Errorf("operation is not a placement")
	}
	child, ok := op.Params["child"].(string)
	if !ok {
		return "", model.IdentityPlacement, fmt.Errorf("child not found or invalid")
	}
	placement, err := extractPlacement(op)
	if err != nil {
		return "", model.IdentityPlacement, err
	}
	return child, placement, nil
}
