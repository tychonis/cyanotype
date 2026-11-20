package hcl

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/core/operation"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) processOperationBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.Operation, error) {
	if len(block.Labels) < 2 {
		return nil, errors.New("invalid block labels")
	}
	cat := block.Labels[0]
	switch cat {
	case "align":
		return p.processAlignOperation(ctx, block)
	default:
		return nil, errors.New("unsupported operation category")
	}
}

func (p *Parser) processAlignOperation(ctx *ParserContext, block *hclsyntax.Block) (*model.Operation, error) {
	name := block.Labels[1]

	content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "points", Required: true},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}

	exprs, diags := hcl.ExprList(content.Attributes["points"].Expr)
	if diags.HasErrors() {
		return nil, diags
	}

	qualifiers := make([]model.Qualifier, 0, len(exprs))
	for _, expr := range exprs {
		ref, err := exprToRef(ctx, expr)
		if err != nil {
			return nil, err
		}
		qualifier := refToQualifier(ctx, ref)
		qualifiers = append(qualifiers, qualifier)
	}

	op := operation.AlignOperation(qualifiers)
	op.Name = name
	op.Qualifier = ctx.NameToQualifier(name)
	return op, nil
}
