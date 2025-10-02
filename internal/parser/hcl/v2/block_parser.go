package hcl

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model/v2"
)

func (c *Core) ParseSymbol(s *UnprocessedSymbol) (sym model.ConcreteSymbol, err error) {
	// Already parsed before.
	if s.qualifier != "" {
		return c.Catalog.Find(s.qualifier)
	}

	// New symbol.
	if s.Block == nil || s.Context == nil {
		return nil, errors.New("illegal nil symbol")
	}
	switch s.Block.Type {
	case "item":
		sym, err = c.parseItemBlock(s.Context, s.Block)
	case "process":
		sym, err = c.parseProcessBlock(s.Context, s.Block)
	case "contract":
		sym, err = c.parseContractBlock(s.Context, s.Block)
	default:
		return nil, errors.New("unknown block type")
	}
	s.qualifier = sym.GetQualifier()
	return
}

func refToQualifier(ctx *ParserContext, ref []string) string {
	if ctx.CurrentModule() == "." {
		return "." + strings.Join(ref, ".")
	} else {
		return strings.Join(append([]string{ctx.CurrentModule()}, ref...), ".")
	}
}

func (c *Core) buildCompanionForItem(ctx *ParserContext, item *model.Item, from []*UnresolvedBOMLine) error {
	slog.Debug("build companions", "module", ctx.CurrentModule(), "item", item.Qualifier)
	var err error
	if len(from) <= 0 {
		return nil
	}
	input := make([]*model.BOMLine, 0, len(from))
	for _, comp := range from {
		qualifier := refToQualifier(ctx, comp.Ref)
		item, err := c.Catalog.Find(qualifier)
		if err != nil {
			if err != catalog.ErrNotFound {
				return err
			} else {
				sym, err := c.Resolve(ctx, comp.Ref)
				if err != nil {
					return err
				}
				unprocessed, ok := sym.(*UnprocessedSymbol)
				if !ok {
					return errors.New("wrong symbol type")
				}
				item, err = c.ParseSymbol(unprocessed)
				if err != nil {
					return err
				}
			}
		}

		compItem, ok := item.(*model.Item)
		if !ok {
			return errors.New("incorrect ref")
		}

		co := &model.CoItem{
			Qualifier: getImplicitCoItemQualifier(compItem),
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
			Qualifier: getImplicitCoProcessQualifier(compItem),
			Input: []*model.BOMLine{
				{
					Item: compItem.Digest,
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
		Qualifier: getImplicitProcessQualifier(item),
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

func (c *Core) parseItemBlock(ctx *ParserContext, block *hclsyntax.Block) (model.ConcreteSymbol, error) {
	item, err := c.blockToItem(ctx, block)
	if err != nil {
		return item, err
	}
	return item, c.Catalog.Add(item)
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

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (model.ConcreteSymbol, error) {
	process, err := c.blockToProcess(ctx, block)
	if err != nil {
		return process, err
	}
	return process, c.Catalog.Add(process)
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

func (c *Core) parseContractBlock(ctx *ParserContext, block *hclsyntax.Block) (model.ConcreteSymbol, error) {
	contract, err := blockToContract(ctx, block)
	if err != nil {
		return contract, err
	}
	return contract, c.Catalog.Add(contract)
}
