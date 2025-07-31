package hcl

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

func (c *Core) parseBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return c.parseImportBlock(ctx, block)
	case "state":
		return c.parseStateBlock(ctx, block)
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

func (c *Core) parseStateBlock(_ *ParserContext, block *hclsyntax.Block) error {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}
	path, err := getString(attrs, "file")
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	b := BOMGraph{}
	err = decoder.Decode(&b)
	if err != nil {
		return err
	}
	c.States[b.RootItem().Qualifier] = &b
	return nil
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
	ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")
	from := readComponents(ctx, attrs["from"])
	return &model.Item{
		Name:       name,
		Qualifier:  ctx.NameToQualifier(name),
		PartNumber: pn,
		Reference:  ref,
		Source:     src,
		From:       from,
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

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	return nil
}

func (c *Core) parseContractBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	return nil
}
