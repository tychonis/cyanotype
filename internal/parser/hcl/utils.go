package hcl

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type ErrorWithRange struct {
	Err   error
	Range *hcl.Range
}

func (e *ErrorWithRange) Error() string {
	return e.Err.Error()
}

func getString(attrs hcl.Attributes, key string) (string, error) {
	attr, ok := attrs[key]
	if !ok {
		return "", errors.New("key not found")
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", diags
	}
	if val.Type() != cty.String {
		return "", errors.New("incorrect type")
	}
	return val.AsString(), nil
}

func getNumber(attrs hcl.Attributes, key string) (float64, error) {
	attr, ok := attrs[key]
	if !ok {
		return 0, errors.New("key not found")
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return 0, diags
	}
	if val.Type() != cty.String {
		return 0, errors.New("incorrect type")
	}
	ret, _ := val.AsBigFloat().Float64()
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

func exprToRef(ctx *ParserContext, expr hclsyntax.Expression) (Ref, error) {
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
