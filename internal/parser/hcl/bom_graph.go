package hcl

import (
	"encoding/json"
	"log/slog"
	"maps"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/states"
	"github.com/tychonis/cyanotype/model"
)

const EXTENSION = ".bpo"

type Items map[string]model.BOMItem

type BOMGraph struct {
	Catalog  *states.Catalog
	Items    Items
	Variants map[string]Items
	Changes  map[string]uuid.UUID
}

func NewBOMGraph() *BOMGraph {
	return &BOMGraph{
		Catalog:  states.NewCatalog(),
		Items:    make(Items),
		Variants: make(map[string]Items),
		Changes:  make(map[string]uuid.UUID),
	}
}

func (g *BOMGraph) MergeGraph(g2 *BOMGraph) error {
	if g2 == nil {
		return nil
	}
	g.Catalog.MergeCatalog(g2.Catalog)
	maps.Copy(g.Items, g2.Items)
	maps.Copy(g.Changes, g2.Changes)
	maps.Copy(g.Variants, g2.Variants)
	return nil
}

func (g *BOMGraph) parseBlock(block *hclsyntax.Block) error {
	switch block.Type {
	case "import":
		return g.parseImportBlock(block)
	case "state":
		return g.parseStateBlock(block)
	case "item":
		return g.parseItemBlock(block)
	}
	return nil
}

func (g *BOMGraph) parseImportBlock(block *hclsyntax.Block) error {
	path := block.Labels[0]
	toImport := parseFolder(path)
	return g.MergeGraph(toImport)
}

func (g *BOMGraph) parseStateBlock(block *hclsyntax.Block) error {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}
	filepath, err := getString(attrs, "file")
	if err != nil {
		return err
	}
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	var c states.Catalog
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return err
	}
	return g.Catalog.MergeCatalog(&c)
}

func blockToItem(block *hclsyntax.Block) (*model.Item, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	pn, _ := getString(attrs, "part_number")
	ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")
	components := readComponents(attrs["from"])
	return &model.Item{
		Name:       name,
		PartNumber: pn,
		Reference:  ref,
		Source:     src,
		Components: components,
	}, nil
}

func (g *BOMGraph) parseItemBlock(block *hclsyntax.Block) error {
	var ok bool
	items := g.Items
	if len(block.Labels) > 1 {
		variant := block.Labels[1]
		items, ok = g.Variants[variant]
		if !ok {
			items = make(Items)
			g.Variants[variant] = items
		}
	}

	item, err := blockToItem(block)
	if err != nil {
		return err
	}

	items[item.Name] = item
	return nil
}

func (g *BOMGraph) assignIDs() {
	for name, item := range g.Items {
		id, ok := g.Catalog.NameIdx[name]
		if !ok {
			id = uuid.New()
			g.Changes[name] = id
		}
		item.SetID(id)
	}
	// TODO: rethink how to handle variants.
	for _, variant := range g.Variants {
		for _, item := range variant {
			if item.GetID() == uuid.Nil {
				item.SetID(uuid.New())
			}
		}
	}
}

func (g *BOMGraph) resolveRefs() {
	for _, item := range g.Items {
		asm, ok := item.(*model.Item)
		if !ok {
			continue
		}
		for _, comp := range asm.Components {
			ref, ok := comp.Item.(*model.SymbolicRef)
			if ok {
				target, found := g.Items[ref.Name]
				if found {
					ref.Target = target
				} else {
					slog.Warn("Unresolved ref.", "target", ref.Name)
				}
			}
		}
	}
}

func (g *BOMGraph) Build() {
	g.assignIDs()
	g.resolveRefs()
}

func parseFile(filename string) *BOMGraph {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		slog.Error("Failed to parse file.", "error", diags.Error())
		return nil
	}

	content, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		slog.Error("Failed to parse content.")
		return nil
	}

	bomGraph := NewBOMGraph()
	for _, block := range content.Blocks {
		bomGraph.parseBlock(block)
	}
	return bomGraph
}

func ParseFile(filename string) *BOMGraph {
	bomGraph := parseFile(filename)
	bomGraph.Build()
	return bomGraph
}

func parseFolder(dir string) *BOMGraph {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	bomGraph := NewBOMGraph()
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == EXTENSION {
			partialBOM := parseFile(filepath.Join(dir, entry.Name()))
			bomGraph.MergeGraph(partialBOM)
		}
	}
	return bomGraph
}

func ParseFolder(dir string) *BOMGraph {
	bomGraph := parseFolder(dir)
	bomGraph.Build()
	return bomGraph
}

func Parse(path string) (*BOMGraph, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var bom *BOMGraph
	if info.IsDir() {
		bom = ParseFolder(path)
	} else {
		bom = ParseFile(path)
	}
	return bom, nil
}
