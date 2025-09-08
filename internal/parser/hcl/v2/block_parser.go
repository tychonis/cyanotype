package hcl

import (
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

func blockToProcess(ctx *ParserContext, block *hclsyntax.Block) (*model.Process, error) {
	name := block.Labels[0]
	_, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	// input := readComponents(ctx, attrs["input"])
	// output := readComponents(ctx, attrs["output"])
	return &model.Process{
		Qualifier: ctx.NameToQualifier(name),
	}, nil
}

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	m := ctx.CurrentModule()
	name := block.Labels[0]
	process, err := blockToProcess(ctx, block)
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
