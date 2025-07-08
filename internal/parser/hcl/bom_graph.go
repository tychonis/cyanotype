package hcl

import (
	"errors"
	"maps"

	"github.com/google/uuid"

	"github.com/tychonis/cyanotype/internal/states"
	"github.com/tychonis/cyanotype/model"
)

const EXTENSION = ".bpo"

type NodeID = uuid.UUID
type ItemID = uuid.UUID

type Items map[ItemID]model.BOMItem
type Nodes map[NodeID]*model.ItemNode

type BOMGraph struct {
	Catalog  *states.Catalog      `json:"catalog" yaml:"catalog"`
	Root     NodeID               `json:"root" yaml:"root"`
	Items    Items                `json:"items" yaml:"items"`
	Nodes    Nodes                `json:"nodes" yaml:"nodes"`
	Usage    map[ItemID][]NodeID  `json:"usage" yaml:"usage"`
	Variants map[string]Items     `json:"-" yaml:"-"`
	Changes  map[string]uuid.UUID `json:"-" yaml:"-"`
}

func NewBOMGraph() *BOMGraph {
	return &BOMGraph{
		Catalog:  states.NewCatalog(),
		Items:    make(Items),
		Nodes:    make(Nodes),
		Usage:    make(map[ItemID][]NodeID),
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

func (g *BOMGraph) assignIDs() {
	for _, item := range g.Items {
		id, ok := g.Catalog.NameIdx[item.GetName()]
		if !ok {
			id = uuid.New()
			g.Changes[item.GetName()] = id
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

func (g *BOMGraph) Build() {
	g.assignIDs()
}

func (g *BOMGraph) AddItem(item *model.Item) error {
	if item == nil {
		return errors.New("nil item")
	}
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
		g.Items[item.ID] = item
	}
	return nil
}

func (g *BOMGraph) AddNode(node *model.ItemNode) error {
	g.Nodes[node.ID] = node
	g.Usage[node.ItemID] = append(g.Usage[node.ItemID], node.ID)
	return nil
}
