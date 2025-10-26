package hcl

import (
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/symbols"
)

func (c *Core) registerBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return c.parseImportBlock(ctx, block)
	default:
		return c.registerUnprocessedBlock(ctx, block)
	}
}

func pathToModuleName(path string) string {
	components := strings.Split(path, "/")
	return components[len(components)-1]
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

func (c *Core) registerUnprocessedBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	name := block.Labels[0]
	symbol := &UnprocessedSymbol{
		Context: ctx,
		Block:   block,
	}
	slog.Debug("Register symbol.", "module", ctx.CurrentModule(), "name", name)
	return c.Symbols.AddSymbol(ctx.CurrentModule(), name, symbol)
}
