package hcl

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/cerror"
	"github.com/tychonis/cyanotype/model"
	"github.com/zclconf/go-cty/cty"
)

type ErrorWithRange struct {
	Err   error
	Range *hcl.Range
}

func (e *ErrorWithRange) Error() string {
	return e.Err.Error()
}

func extractAttributes(body *hclsyntax.Body) (hcl.Attributes, error) {
	attrs := make(hcl.Attributes, len(body.Attributes))
	for name, attr := range body.Attributes {
		attrs[name] = attr.AsHCLAttribute()
	}
	return attrs, nil
}

func getString(attrs hcl.Attributes, key string) (string, error) {
	attr, ok := attrs[key]
	if !ok {
		return "", cerror.KeyNotFoundError(key)
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", diags
	}
	if val.Type() != cty.String {
		return "", errors.New("incorrect type for " + key)
	}
	return val.AsString(), nil
}

func getNumber(attrs hcl.Attributes, key string) (float64, error) {
	attr, ok := attrs[key]
	if !ok {
		return 0, cerror.KeyNotFoundError(key)
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return 0, diags
	}
	if val.Type() != cty.Number {
		return 0, errors.New("incorrect type for " + key)
	}
	ret, _ := val.AsBigFloat().Float64()
	return ret, nil
}

func getFloat64Array(attrs hcl.Attributes, key string) ([]float64, error) {
	attr, ok := attrs[key]
	if !ok {
		return nil, cerror.KeyNotFoundError(key)
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return nil, diags
	}
	if !val.Type().IsTupleType() && !val.Type().IsListType() {
		return nil, errors.New("incorrect type for " + key)
	}
	slice := val.AsValueSlice()
	ret := make([]float64, 0, len(slice))
	for _, num := range slice {
		f, _ := num.AsBigFloat().Float64()
		ret = append(ret, f)
	}
	return ret, nil
}

func getPlacement(attrs hcl.Attributes, key string) (model.Placement, error) {
	attr, ok := attrs[key]
	if !ok {
		return model.Placement{}, cerror.KeyNotFoundError(key)
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return model.Placement{}, diags
	}
	return parsePlacement(val)
}

func parsePlacement(val cty.Value) (model.Placement, error) {
	var ret model.Placement

	if !val.Type().IsTupleType() && !val.Type().IsListType() {
		return ret, errors.New("expected tuple or list")
	}

	if val.IsNull() || !val.IsKnown() {
		return ret, errors.New("invalid placement value")
	}

	slice := val.AsValueSlice()
	if len(slice) != 7 {
		return ret, errors.New("expected 7 numbers")
	}

	for i, num := range slice {
		if !num.Type().Equals(cty.Number) || num.IsNull() || !num.IsKnown() {
			return ret, fmt.Errorf("expected number at index %d", i)
		}

		v, _ := num.AsBigFloat().Float64()
		if i < 4 {
			ret.Rotation[i] = v
		} else {
			ret.Position[i-4] = v
		}
	}

	return ret, nil
}

func getObjectKey(expr hcl.Expression) string {
	key, ok := expr.(*hclsyntax.ObjectConsKeyExpr)
	if !ok {
		return ""
	}
	val, diags := key.Value(nil)
	if diags.HasErrors() {
		return ""
	}
	return val.AsString()
}

func getTraverserName(t hcl.Traverser) string {
	switch t := t.(type) {
	case hcl.TraverseRoot:
		return t.Name
	case hcl.TraverseAttr:
		return t.Name
	default:
		return ""
	}
}

func exprToRef(ctx *ParserContext, expr hcl.Expression) (Ref, error) {
	se, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return nil, errors.New("incorrect expr type")
	}
	ref := make(Ref, 0)
	if ctx.CurrentModule() != "." {
		ref = append(ref, ctx.CurrentModule())
	}
	for _, n := range se.Traversal {
		ref = append(ref, getTraverserName(n))
	}
	return ref, nil
}
