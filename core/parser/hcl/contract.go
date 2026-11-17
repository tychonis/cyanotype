package hcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) getContractParams(ctx *ParserContext, cat string, attrs hcl.Attributes) (map[string]any, error) {
	return nil, nil
}

func (p *Parser) getImplicitOperationParams(ctx *ParserContext, cat string, attrs hcl.Attributes) (map[string]any, error) {
	params := make(map[string]any)
	if cat == "point" {
		location, err := getPlacement(attrs, "location")
		if err != nil {
			return nil, err
		}
		params["location"] = location
	}
	return params, nil
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

func (p *Parser) processImplicitOperation(ctx *ParserContext, block *hclsyntax.Block) (*model.Operation, error) {
	cat := block.Labels[0]
	name := block.Labels[1]

	if !p.hasImplicitOperation(ctx, cat) {
		return nil, nil
	}
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	params, err := p.getImplicitOperationParams(ctx, cat, attrs)
	if err != nil {
		return nil, err
	}
	op := &model.Operation{
		Category: cat,
		Name:     name,
		Params:   params,
	}
	op.Qualifier = ctx.NameToQualifier(name)
	digest, err := digest.SHA256FromSymbol(op)
	if err != nil {
		return op, err
	}
	op.Digest = digest
	return op, p.Symbols.RegisterConcreteSymbol(op)
}
