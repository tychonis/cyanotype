package hcl

import (
	"errors"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/model"
)

type Ref = []string

type UnresolvedBOMLine struct {
	Role string  `json:"role" yaml:"role"`
	Ref  Ref     `json:"ref" yaml:"ref"`
	Qty  float64 `json:"qty" yaml:"qty"`
}

func readBOMLine(ctx *ParserContext, obj *hclsyntax.ObjectConsExpr) *UnresolvedBOMLine {
	ret := &UnresolvedBOMLine{
		Qty: 1,
	}
	for _, item := range obj.Items {
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

func parseBOMLineAttr(ctx *ParserContext, attr *hcl.Attribute) []*UnresolvedBOMLine {
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
		qualifier := refToQualifier(ctx, comp.Ref)
		compItemSym, err := c.Catalog.Find(qualifier)
		if err != nil {
			if err != catalog.ErrNotFound {
				return nil, err
			} else {
				sym, err := c.Resolve(ctx, comp.Ref)
				if err != nil {
					return nil, err
				}
				unprocessed, ok := sym.(*UnprocessedSymbol)
				if !ok {
					return nil, errors.New("wrong symbol type")
				}
				compItemSym, err = c.ParseSymbol(unprocessed)
				if err != nil {
					return nil, err
				}
			}
		}

		compItem, ok := compItemSym.(*model.Item)
		if !ok {
			return nil, errors.New("incorrect ref")
		}
		compCoItems, err := c.Catalog.GetCoItems(compItem.Digest)
		if err != nil {
			return nil, err
		}
		if len(compCoItems) != 1 {
			slog.Debug("error", "item", compItem.Qualifier, "length", len(compCoItems), "digest", compItem.Digest)
			return nil, errors.New("not implemented yet")
		}
		ret = append(ret, &model.BOMLine{
			Item: compCoItems[0].Item,
			Qty:  comp.Qty,
		})
	}
	return ret, nil
}

func (c *Core) processKeywordIMPL(ctx *ParserContext, impl []Ref) ([]*model.Contract, error) {
	return nil, nil
}
