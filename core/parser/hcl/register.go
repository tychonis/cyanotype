package hcl

import (
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/symbols"
)

func (p *Parser) registerBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return p.parseImportBlock(ctx, block)
	default:
		return p.registerUnprocessedBlock(ctx, block)
	}
}

func pathToModuleName(path string) string {
	components := strings.Split(path, "/")
	return components[len(components)-1]
}

func (p *Parser) parseImportBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	path := block.Labels[0]
	moduleName := pathToModuleName(path)
	currentModule := ctx.CurrentModule()
	err := p.Symbols.AddSymbol(currentModule, moduleName,
		&symbols.Import{Symbols: p.Symbols, Identifier: path})
	if err != nil {
		return err
	}
	newCtx, err := ctx.Import(path)
	if err != nil {
		return err
	}
	return p.parseFolder(newCtx, path)
}

func (p *Parser) registerUnprocessedBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	name := block.Labels[0]
	symbol := &UnprocessedSymbol{
		Context: ctx,
		Block:   block,
	}
	slog.Debug("Register symbol.", "module", ctx.CurrentModule(), "name", name)
	return p.Symbols.AddSymbol(ctx.CurrentModule(), name, symbol)
}
