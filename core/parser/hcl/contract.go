package hcl

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/core/operation"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) getContractParams(ctx *ParserContext, cat string, attrs hcl.Attributes) (map[string]any, error) {
	return nil, nil
}

func (p *Parser) getImplicitOperation(ctx *ParserContext, cat string, attrs hcl.Attributes) (*model.Operation, error) {
	if cat == "point" {
		location, err := getPlacement(attrs, "location")
		if err != nil {
			return nil, err
		}
		return operation.PointOperation(location), nil
	}
	return nil, errors.ErrUnsupported
}

func (p *Parser) processContractBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.Contract, error) {
	cat := block.Labels[0]
	name := block.Labels[1]

	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	params, err := p.getContractParams(ctx, cat, attrs)
	if err != nil {
		return nil, err
	}
	contract := &model.Contract{
		Category: cat,
		Name:     name,
		Params:   params,
	}
	contract.Qualifier = ctx.NameToQualifier(name)
	digest, err := digest.SHA256FromSymbol(contract)
	if err != nil {
		return contract, err
	}
	contract.Digest = digest
	return contract, p.Symbols.RegisterConcreteSymbol(contract)
}

func (p *Parser) hasImplicitOperation(ctx *ParserContext, cat string) bool {
	if cat == "point" {
		return true
	}
	return false
}

func (p *Parser) generateImplicitOperation(ctx *ParserContext, block *hclsyntax.Block) (*model.Operation, error) {
	if len(block.Labels) < 2 {
		return nil, errors.New("invalid block labels")
	}
	cat := block.Labels[0]
	name := block.Labels[1]

	if !p.hasImplicitOperation(ctx, cat) {
		return nil, nil
	}
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	op, err := p.getImplicitOperation(ctx, cat, attrs)
	if err != nil {
		return nil, err
	}
	op.Name = name
	op.Qualifier = ctx.NameToQualifier(name)
	return op, nil
}
