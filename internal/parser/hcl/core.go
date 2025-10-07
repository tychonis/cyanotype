package hcl

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

const EXTENSION string = ".bpo"
const IMPLICIT string = "implicit."
const DEFAULT string = "default"

type Core struct {
	Symbols *symbols.SymbolTable
	Catalog catalog.Catalog
}

type ParserContext struct {
	ImportStack []string
}

func NewParserContext() *ParserContext {
	return &ParserContext{
		ImportStack: []string{"."},
	}
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

func (ctx *ParserContext) NameToQualifier(name string) string {
	prefix := ctx.CurrentModule()
	if prefix == "." {
		prefix = ""
	}
	return prefix + "." + name
}

func NewCore(catalogType string) *Core {
	return &Core{
		Symbols: symbols.NewSymbolTable(),
		Catalog: catalog.NewCatalog(catalogType),
	}
}

func (c *Core) Resolve(ctx *ParserContext, ref []string) (model.Symbol, error) {
	slog.Debug("Resolving ref", "module", ctx.CurrentModule(), "ref", ref)
	mod, ok := c.Symbols.Modules[ctx.CurrentModule()]
	if !ok {
		return nil, errors.New("no registered symbols")
	}
	return mod.Resolve(ref)
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
	ctx := NewParserContext()
	return c.parseFolder(ctx, dir)
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
		err := c.registerBlock(ctx, block)
		if err != nil {
			slog.Warn("Error parsing block.", "error", err)
		}
	}
	return nil
}

func (c *Core) ParseFile(filename string) error {
	ctx := NewParserContext()
	return c.parseFile(ctx, filename)
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

func (c *Core) Process(path string) error {
	err := c.Parse(path)
	if err != nil {
		return err
	}
	return c.processModules(c.Catalog)
}

func (c *Core) processModules(cat catalog.Catalog) error {
	for _, module := range c.Symbols.Modules {
		err := c.processModuleScope(module, cat)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Core) processModuleScope(m *symbols.ModuleScope, cat catalog.Catalog) error {
	for _, symbol := range m.Symbols {
		switch s := symbol.(type) {
		case *symbols.ModuleScope:
			c.processModuleScope(s, cat)
		case *UnprocessedSymbol:
			slog.Debug("Process item", "item", s)
			_, err := c.ParseSymbol(s)
			if err != nil {
				slog.Warn("Error adding item.", "error", err, "item", s)
			}
		default:
		}
	}
	return nil
}

func (c *Core) ResolveBOMLine(ctx *ParserContext, line *UnresolvedBOMLine) (*model.BOMLine, error) {
	s, err := c.Resolve(ctx, line.Ref)
	if err != nil {
		return nil, err
	}
	item, ok := s.(*model.Item)
	if !ok {
		return nil, errors.New("ref is not an item")
	}
	return &model.BOMLine{
		Role: line.Role,
		Item: item.Digest,
		Qty:  line.Qty,
	}, nil
}
