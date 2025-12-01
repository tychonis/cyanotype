package hcl

import (
	"errors"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/model"
)

type Ref = []string

type UnresolvedBOMLine struct {
	Role string  `json:"role" yaml:"role"`
	Ref  Ref     `json:"ref" yaml:"ref"`
	Qty  float64 `json:"qty" yaml:"qty"`
}

func readBOMLine(ctx *ParserContext, expr *hclsyntax.ObjectConsExpr) *UnresolvedBOMLine {
	ret := &UnresolvedBOMLine{
		Qty: 1,
	}
	for _, item := range expr.Items {
		key := getObjectKey(item.KeyExpr)
		switch key {
		case "role":
			val, _ := item.ValueExpr.Value(nil)
			ret.Role = val.AsString()
		case "ref":
			ref, _ := exprToRef(ctx, item.ValueExpr)
			ret.Ref = ref
		case "qty":
			val, _ := item.ValueExpr.Value(nil)
			ret.Qty, _ = val.AsBigFloat().Float64()
		}
	}
	return ret
}

func parseBOMLinesAttr(ctx *ParserContext, attr *hcl.Attribute) []*UnresolvedBOMLine {
	if attr == nil {
		return nil
	}

	expr, ok := attr.Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil
	}

	comps := make([]*UnresolvedBOMLine, 0)
	for _, elem := range expr.Exprs {
		obj, ok := elem.(*hclsyntax.ObjectConsExpr)
		if !ok {
			continue
		}
		comp := readBOMLine(ctx, obj)
		comps = append(comps, comp)
	}
	return comps
}

func (c *Core) processKeywordFROM(ctx *ParserContext, from []*UnresolvedBOMLine) ([]*model.BOMLine, error) {
	if len(from) <= 0 {
		return nil, nil
	}
	ret := make([]*model.BOMLine, 0, len(from))
	for _, comp := range from {
		item, err := c.resolveBOMLineRef(ctx, comp.Ref)
		if err != nil {
			return nil, err
		}
		coItems, err := c.Catalog.GetCoItems(item.Digest)
		if err != nil {
			return nil, err
		}
		if len(coItems) != 1 {
			slog.Debug("error", "item", item.Qualifier, "length", len(coItems), "digest", item.Digest)
			return nil, errors.New("not implemented yet")
		}
		ret = append(ret, &model.BOMLine{
			Item: coItems[0].Item,
			Qty:  comp.Qty,
		})
	}
	return ret, nil
}

func (c *Core) readContractLine(ctx *ParserContext, expr hcl.Expression) ([]Ref, error) {
	ret := make([]Ref, 0)
	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		for _, el := range e.Exprs {
			ref, err := exprToRef(ctx, el)
			if err != nil {
				return nil, err
			}
			ret = append(ret, ref)
		}
	default:
		ref, err := exprToRef(ctx, e)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ref)
	}
	return ret, nil
}

func (c *Core) resolveContractsID(ctx *ParserContext, contracts []Ref) ([]model.ContractID, error) {
	ret := make([]model.ContractID, 0, len(contracts))
	for _, ref := range contracts {
		sym, err := c.Resolve(ctx, ref)
		if err != nil {
			return nil, err
		}
		contract, ok := sym.(*model.Contract)
		if !ok {
			return nil, errors.New("implement non contract")
		}
		ret = append(ret, contract.Digest)
	}
	return ret, nil
}
