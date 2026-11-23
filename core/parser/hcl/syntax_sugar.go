package hcl

import (
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/core/operation"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/model"
)

type Ref = []string

type UnresolvedBOMLine struct {
	Name         string          `json:"name" yaml:"name"`
	Ref          Ref             `json:"ref" yaml:"ref"`
	Qty          float64         `json:"qty" yaml:"qty"`
	HasPlacement bool            `json:"-" yaml:"-"`
	Placement    model.Placement `json:"placement,omitempty" yaml:"placement,omitempty"`
}

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
			val, diags := item.ValueExpr.Value(nil)
			if diags.HasErrors() {
				slog.Warn("error reading placement value", "err", diags)
				continue
			}
			placement, err := parsePlacement(val)
			if err != nil {
				slog.Warn("error parsing placement", "err", err)
				continue
			}
			ret.Placement = placement
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

func (p *Parser) resolveBOMLines(ctx *ParserContext, proc *model.Process, from []*UnresolvedBOMLine) ([]*model.Operation, error) {
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

	operations := make([]*model.Operation, 0, len(from))
	if drawing {
		proc.Set("purpose", "drawing")
	}
	proc.Input = make([]*model.BOMLine, 0, len(from))

	for _, comp := range from {
		item, err := p.resolveBOMLineRef(ctx, comp.Ref)
		if err != nil {
			return operations, err
		}
		// Since this is a syntax sugar for keyword FROM, we will only use the
		// companion coitem generated for the item. At this point, because we
		// already run c.resolveBOMLineRef, the companion coitem of child item
		// should already be generated and registered in the symbol table.
		coItemQualifier := qualifier.ImplicitCoItem(item)
		coItemSym, err := p.Symbols.FindConcreteSymbol(coItemQualifier)
		if err != nil {
			return operations, err
		}
		proc.Input = append(proc.Input, &model.BOMLine{
			Name: comp.Name,
			Item: coItemSym.GetDigest(),
			Qty:  comp.Qty,
		})
		if comp.HasPlacement {
			op := operation.PlacementOperation(comp.Name, comp.Placement)
			op.Name = "_" + comp.Name
			op.Qualifier = ctx.NameToQualifier(op.Name)
			operations = append(operations, op)
		}
	}
	return operations, nil
}
