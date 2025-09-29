package hcl

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model/v2"
)

func (c *Core) ParseSymbol(s *UnprocessedSymbol) (model.Symbol, error) {
	if s.Block == nil || s.Context == nil {
		return nil, errors.New("illegal nil symbol")
	}
	switch s.Block.Type {
	case "item":
		return c.parseItemBlock(s.Context, s.Block)
	case "process":
		return c.parseProcessBlock(s.Context, s.Block)
	case "contract":
		return c.parseContractBlock(s.Context, s.Block)
	default:
		return nil, errors.New("unknown block type")
	}
}

func (c *Core) buildCompanionForItem(ctx *ParserContext, item *model.Item, from []*UnresolvedBOMLine) error {
	var err error
	if len(from) <= 0 {
		return nil
	}
	input := make([]*model.BOMLine, 0, len(from))
	for _, comp := range from {
		ref, err := c.Resolve(ctx, comp.Ref)
		if err != nil {
			return err
		}
		var itemRef *model.Item
		switch s := ref.(type) {
		case *UnprocessedSymbol:
			itemSymbol, err := c.ParseSymbol(s)
			if err != nil {
				return err
			}
			item, ok := itemSymbol.(*model.Item)
			if !ok {
				return errors.New("incorrect ref")
			}
			itemRef = item
		case *model.Item:
			itemRef = s
		default:
			return errors.New("incorrect ref")
		}
		co := &model.CoItem{
			Qualifier: IMPLICIT + itemRef.Qualifier + ".coitem",
		}
		co.Digest, err = digest.SHA256FromSymbol(co)
		if err != nil {
			return err
		}
		input = append(input, &model.BOMLine{
			Item: co.Digest,
			Qty:  comp.Qty,
		})

		cp := &model.CoProcess{
			Qualifier: IMPLICIT + itemRef.Qualifier + ".coprocess",
			Input: []*model.BOMLine{
				{
					Item: itemRef.Digest,
					Qty:  1,
				},
			},
			Output: []*model.BOMLine{
				{
					Item: co.Digest,
					Qty:  1,
				},
			},
		}
		cp.Digest, err = digest.SHA256FromSymbol(cp)
		if err != nil {
			return err
		}
	}
	p := &model.Process{
		Qualifier: IMPLICIT + item.Qualifier + ".process",
		Output: []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
				Role: DEFAULT,
			},
		},
		Input: input,
	}
	p.Digest, err = digest.SHA256FromSymbol(p)
	return err
}

func (c *Core) blockToItem(ctx *ParserContext, block *hclsyntax.Block) (*model.Item, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	pn, _ := getString(attrs, "part_number")
	// ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")
	from := readComponents(ctx, attrs["from"])
	item := &model.Item{
		Qualifier: ctx.NameToQualifier(name),
		Content: &model.ItemContent{
			Name:       name,
			Source:     src,
			PartNumber: pn,
		},
	}
	var err error
	item.Digest, err = digest.SHA256FromSymbol(item)
	if err != nil {
		return item, err
	}
	err = c.buildCompanionForItem(ctx, item, from)
	return item, err
}

func (c *Core) parseItemBlock(ctx *ParserContext, block *hclsyntax.Block) (model.Symbol, error) {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	item, err := c.blockToItem(ctx, block)
	if err != nil {
		return item, err
	}
	return item, c.Symbols.UpdateSymbol(m, name, item)
}

func (c *Core) createProcessContract(process *model.Process, mode string, line *UnresolvedBOMLine) (*model.Contract, error) {
	ret := &model.Contract{
		Qualifier: fmt.Sprintf("%s.%s.%s", process.Qualifier, mode, line.Role),
	}
	return ret, nil
}

func (c *Core) blockToProcess(ctx *ParserContext, block *hclsyntax.Block) (*model.Process, error) {
	name := block.Labels[0]
	ret := &model.Process{
		Qualifier: ctx.NameToQualifier(name),
	}
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	cycle, _ := getNumber(attrs, "cycle")
	ret.CycleTime = cycle
	input := readComponents(ctx, attrs["input"])
	ret.Input = make([]*model.BOMLine, 0, len(input))
	for _, line := range input {
		resolved, err := c.ResolveBOMLine(ctx, line)
		if err != nil {
			return nil, err
		}
		ret.Input = append(ret.Input, resolved)
	}
	output := readComponents(ctx, attrs["output"])
	ret.Output = make([]*model.BOMLine, 0, len(output))
	for _, line := range output {
		resolved, err := c.ResolveBOMLine(ctx, line)
		if err != nil {
			return nil, err
		}
		ret.Output = append(ret.Output, resolved)
	}
	return ret, nil
}

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (model.Symbol, error) {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	process, err := c.blockToProcess(ctx, block)
	if err != nil {
		return process, err
	}
	return process, c.Symbols.UpdateSymbol(m, name, process)
}

func blockToContract(ctx *ParserContext, block *hclsyntax.Block) (*model.Contract, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	params := make(map[string]any)
	for attr := range attrs {
		// TODO: support more types?
		param, _ := getString(attrs, attr)
		params[attr] = param
	}
	return &model.Contract{
		Name:      name,
		Qualifier: ctx.NameToQualifier(name),
		Params:    params,
	}, nil
}

func (c *Core) parseContractBlock(ctx *ParserContext, block *hclsyntax.Block) (model.Symbol, error) {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	contract, err := blockToContract(ctx, block)
	if err != nil {
		return contract, err
	}
	return contract, c.Symbols.UpdateSymbol(m, name, contract)
}
