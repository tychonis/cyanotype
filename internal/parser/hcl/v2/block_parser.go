package hcl

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model/v2"
)

func (c *Core) parseBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return c.parseImportBlock(ctx, block)
	case "item":
		return c.parseItemBlock(ctx, block)
	case "process":
		return c.parseProcessBlock(ctx, block)
	case "contract":
		return c.parseContractBlock(ctx, block)
	}
	return nil
}

func (c *Core) parseImportBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	path := block.Labels[0]
	moduleName := pathToModuleName(path)
	currentModule := ctx.CurrentModule()
	err := c.Symbols.AddSymbol(currentModule, moduleName,
		&symbols.Import{Symbols: c.Symbols, Identifier: path})
	if err != nil {
		return err
	}
	newCtx, err := ctx.Import(path)
	if err != nil {
		return err
	}
	return c.parseFolder(newCtx, path)
}

func pathToModuleName(path string) string {
	components := strings.Split(path, "/")
	return components[len(components)-1]
}

func blockToItem(ctx *ParserContext, block *hclsyntax.Block) (*model.Item, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	pn, _ := getString(attrs, "part_number")
	// ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")
	// from := readComponents(ctx, attrs["from"])
	return &model.Item{
		Qualifier: ctx.NameToQualifier(name),
		Content: &model.ItemContent{
			Name:       name,
			PartNumber: pn,
			Source:     src,
		},
	}, nil
}

func (c *Core) parseItemBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	item, err := blockToItem(ctx, block)
	if err != nil {
		return err
	}
	return c.Symbols.AddSymbol(m, name, item)
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
	// input := readComponents(ctx, attrs["input"])
	// ret.Input = make([]*model.BOMLine, 0, len(input))
	// for _, line := range input {
	// 	resolved, err := c.ResolveBOMLine(ctx, line)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	ret.Input = append(ret.Input, resolved)
	// }
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

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	process, err := c.blockToProcess(ctx, block)
	if err != nil {
		return err
	}
	return c.Symbols.AddSymbol(m, name, process)
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

func (c *Core) parseContractBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	contract, err := blockToContract(ctx, block)
	if err != nil {
		return err
	}
	return c.Symbols.AddSymbol(m, name, contract)
}
