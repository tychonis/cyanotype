package hcl

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/tychonis/cyanotype/model"
)

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

func readComponent(ctx *ParserContext, obj *hclsyntax.ObjectConsExpr) *model.Component {
	ret := &model.Component{
		Qty: 1,
	}
	for _, item := range obj.Items {
		key := getObjectKey(item.KeyExpr)
		switch key {
		case "name":
			val, _ := item.ValueExpr.Value(nil)
			ret.Name = val.AsString()
		case "ref":
			se, ok := item.ValueExpr.(*hclsyntax.ScopeTraversalExpr)
			if !ok {
				return nil
			}
			ref := make([]string, 0)
			if ctx.CurrentModule() != "." {
				ref = append(ref, ctx.CurrentModule())
			}
			for _, n := range se.Traversal {
				ref = append(ref, getTraverserName(n))
			}
			ret.Ref = ref
		case "qty":
			val, _ := item.ValueExpr.Value(nil)
			ret.Qty, _ = val.AsBigFloat().Float64()
		}
	}
	return ret
}

func readComponents(ctx *ParserContext, attr *hcl.Attribute) []*model.Component {
	if attr == nil {
		return nil
	}

	expr, ok := attr.Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	comps := make([]*model.Component, 0)
	for _, elem := range expr.Exprs {
		obj, ok := elem.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}
		comp := readComponent(ctx, obj)
		comps = append(comps, comp)
	}
	return comps
}
