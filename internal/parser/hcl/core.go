package hcl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

type Core struct {
	Symbols *symbols.SymbolTable
	States  map[NodeID]*BOMGraph
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
		States:  make(map[NodeID]*BOMGraph),
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
	case "state":
		err := c.parseStateBlock(ctx, block)
		if err != nil {
			fmt.Printf("%+v", err)
		}
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
	c.States[b.Root] = &b
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
	bomGraph.Root = rootNode.ID
	bomGraph.AddNode(rootNode)
	for _, comp := range rootItem.GetComponents() {
		c.buildBom(bomGraph, comp.Name, comp.Ref, rootNode, comp.Qty)
	}
	bomGraph.BuildCatalog()

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
