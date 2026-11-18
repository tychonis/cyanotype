package operation

import (
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

func PlacementOperation(child model.ConcreteSymbol, p model.Placement) *model.Operation {
	op := &model.Operation{
		Category: "placement",
		Params:   make(map[string]interface{}),
	}
	op.Params["placement"] = p
	op.Params["child"] = child.GetQualifier()
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
	placement, ok := op.Params["placement"].(model.Placement)
	if !ok {
		return "", model.IdentityPlacement, fmt.Errorf("placement not found or invalid")
	}
	return child, placement, nil
}
