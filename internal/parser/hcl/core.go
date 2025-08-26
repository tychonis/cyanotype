package hcl

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

type Core struct {
	Symbols *symbols.SymbolTable
	States  map[string]*BOMGraph
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

func (ctx *ParserContext) NameToQualifier(name string) string {
	prefix := ctx.CurrentModule()
	if prefix == "." {
		prefix = ""
	}
	return prefix + "." + name
}

func NewCore() *Core {
	return &Core{
		Symbols: symbols.NewSymbolTable(),
		States:  make(map[string]*BOMGraph),
	}
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
		err := c.parseBlock(ctx, block)
		if err != nil {
			slog.Warn("Error parsing block.", "error", err)
		}
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

func (c *Core) Build(path string, root []string) (*BOMGraph, error) {
	// TODO: check parsed.
	bomGraph := NewBOMGraph()
	rootSymbol, err := c.Symbols.Resolve(root)
	if err != nil {
		return bomGraph, nil
	}
	rootItem, ok := rootSymbol.(*model.Item)
	if !ok {
		return bomGraph, errors.New("unrecongnized root")
	}
	bomGraph.AddItem(rootItem)
	rootNode := &model.ItemNode{
		ID:       uuid.New(),
		ItemID:   rootItem.ID,
		Path:     "/" + "root",
		Children: make([]NodeID, 0),
		Qty:      1,
	}
	bomGraph.Roots = []NodeID{rootNode.ID}
	bomGraph.AddNode(rootNode)
	for _, comp := range rootItem.GetComponents() {
		c.buildBom(bomGraph, comp.Name, comp.Ref, rootNode, comp.Qty)
	}
	bomGraph.BuildIndex()

	return bomGraph, nil
}

func (c *Core) buildBom(bom *BOMGraph, name string, ref []string, parent *model.ItemNode, qty float64) {
	symbol, err := c.Symbols.Resolve(ref)
	if err != nil {
		return
	}
	item, ok := symbol.(*model.Item)
	if !ok {
		return
	}
	bom.AddItem(item)
	node := &model.ItemNode{
		ID:       uuid.New(),
		ItemID:   item.ID,
		Path:     parent.Path + "/" + name,
		ParentID: parent.ID,
		Children: make([]NodeID, 0),
		Qty:      qty,
	}
	bom.AddNode(node)
	parent.Children = append(parent.Children, node.ID)
	for _, comp := range item.GetComponents() {
		c.buildBom(bom, comp.Name, comp.Ref, node, comp.Qty)
	}
}
