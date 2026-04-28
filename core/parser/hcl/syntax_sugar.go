package hcl

import (
	"errors"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

type Ref = []string

type UnresolvedBOMLine struct {
	Name         string  `json:"name" yaml:"name"`
	Ref          Ref     `json:"ref" yaml:"ref"`
	Qty          float64 `json:"qty" yaml:"qty"`
	HasPlacement bool
	Placement    [7]float64 `json:"placement" yaml:"placement"`
}

var IdentityPlacement = [7]float64{0, 0, 0, 1, 0, 0, 0}

func readBOMLine(ctx *ParserContext, expr *hclsyntax.ObjectConsExpr) *UnresolvedBOMLine {
	ret := &UnresolvedBOMLine{
		Qty:          1,
		HasPlacement: false,
	}
	for _, item := range expr.Items {
		key := getObjectKey(item.KeyExpr)
		switch key {
		case "name":
			val, _ := item.ValueExpr.Value(nil)
			ret.Name = val.AsString()
		case "ref":
			ref, _ := exprToRef(ctx, item.ValueExpr)
			ret.Ref = ref
		case "qty":
			val, _ := item.ValueExpr.Value(nil)
			ret.Qty, _ = val.AsBigFloat().Float64()
		case "placement":
			ret.HasPlacement = true
			val, _ := item.ValueExpr.Value(nil)
			slice := val.AsValueSlice()
			if len(slice) != 7 {
				slog.Warn("incorrect format for placement")
				continue
			}
			for i, num := range slice {
				ret.Placement[i], _ = num.AsBigFloat().Float64()
			}
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

func (c *Core) processKeywordFROM(ctx *ParserContext, from []*UnresolvedBOMLine) (process.ProcessContent, error) {
	if len(from) <= 0 {
		return nil, nil
	}
	drawing := false
	for _, comp := range from {
		if comp.HasPlacement {
			drawing = true
			break
		}
	}

	components := make([]*process.Component, 0, len(from))
	input := make([]*model.BOMLine, 0, len(from))

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
			slog.Debug("error getting coitems", "item", item.Qualifier, "length", len(coItems), "digest", item.Digest)
			return nil, errors.New("non-unique coitems not implemented yet")
		}
		if drawing {
			if !comp.HasPlacement {
				slog.Warn("component has no placement for drawing", "component", comp.Name, "ref", comp.Ref)
				comp.Placement = IdentityPlacement
			}
			components = append(components, &process.Component{
				Name:        comp.Name,
				CoItem:      coItems[0].Item,
				Rotation:    [4]float64(comp.Placement[:4]),
				Translation: [3]float64(comp.Placement[4:]),
			})
		} else {
			input = append(input, &model.BOMLine{
				Name: comp.Name,
				Item: coItems[0].Item,
				Qty:  comp.Qty,
			})
		}
	}
	var ret process.ProcessContent
	if drawing {
		ret = &process.Drawing{
			Components: components,
		}
	} else {
		ret = &process.Abstract{
			Input: input,
		}
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
