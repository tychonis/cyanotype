package hcl

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

type Core struct {
	Symbols *symbols.SymbolTable
}

type ParserContext struct {
	ImportStack []string
}

func (ctx *ParserContext) Import(path string) (*ParserContext, error) {
	for _, existed := range ctx.ImportStack {
		if existed == path {
			return nil, errors.New("cyclic import detected")
		}
	}
	return &ParserContext{
		ImportStack: append([]string{path}, ctx.ImportStack...),
	}, nil
}

func (ctx *ParserContext) CurrentModule() string {
	return ctx.ImportStack[0]
}

func NewCore() *Core {
	return &Core{Symbols: symbols.NewSymbolTable()}
}

func (c *Core) parseFolder(ctx *ParserContext, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == EXTENSION {
			err = c.parseFile(ctx, filepath.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Core) ParseFolder(dir string) error {
	ctx := ParserContext{
		ImportStack: []string{"."},
	}
	return c.parseFolder(&ctx, dir)
}

func (c *Core) parseFile(ctx *ParserContext, filename string) error {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		slog.Error("Failed to parse file.", "error", diags.Error())
		return diags
	}

	content, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		slog.Error("Failed to parse content.")
		return errors.New("failed to parse content")
	}

	for _, block := range content.Blocks {
		c.parseBlock(ctx, block)
	}
	return nil
}

func (c *Core) ParseFile(filename string) error {
	ctx := ParserContext{
		ImportStack: []string{"."},
	}
	return c.parseFile(&ctx, filename)
}

func (c *Core) Parse(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return c.ParseFolder(path)
	}
	return c.ParseFile(path)
}

func (c *Core) parseBlock(ctx *ParserContext, block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return c.parseImportBlock(ctx, block)
	// case "state":
	// 	return c.parseStateBlock(block)
	case "item":
		return c.parseItemBlock(ctx, block)
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
	ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")
	from := readComponents(ctx, attrs["from"])
	return &model.Item{
		Name:       name,
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

func (c *Core) Build(path string) (*BOMGraph, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		c.ParseFolder(path)
	} else {
		c.ParseFile(path)
	}
	bomGraph := NewBOMGraph()
	for moduleName, module := range c.Symbols.Modules {
		for symbolName, symbol := range module.Symbols {
			item, ok := symbol.(model.BOMItem)
			if !ok {
				continue
			}
			fullName := moduleName + "." + symbolName
			bomGraph.Items[fullName] = item
		}
	}
	bomGraph.Build()
	return bomGraph, nil
}
