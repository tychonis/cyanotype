package hcl

import (
	"errors"
	"maps"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/tychonis/cyanotype/internal/match"
	"github.com/tychonis/cyanotype/model"
)

const EXTENSION = ".bpo"

type NodeID = uuid.UUID
type ItemID = uuid.UUID

type Items map[ItemID]*model.Item
type Nodes map[NodeID]*model.ItemNode

type BOMGraph struct {
	Root  NodeID              `json:"root" yaml:"root"`
	Items Items               `json:"items" yaml:"items"`
	Nodes Nodes               `json:"nodes" yaml:"nodes"`
	Usage map[ItemID][]NodeID `json:"usage" yaml:"usage"`

	Variants map[string]Items     `json:"-" yaml:"-"`
	Changes  map[string]uuid.UUID `json:"-" yaml:"-"`

	ID      uuid.UUID `json:"id" yaml:"id"`
	Version string    `json:"version" yaml:"version"`

	QualifierIndex  map[string]ItemID `json:"qualifier_index" yaml:"qualifier_index"`
	PartNumberIndex map[string]ItemID `json:"part_number_index" yaml:"part_number_index"`
	PathIndex       map[string]NodeID `json:"path_index" yaml:"path_index"`
}

func NewBOMGraph() *BOMGraph {
	return &BOMGraph{
		ID:      uuid.New(),
		Version: "alpha-0",

		Items:    make(Items),
		Nodes:    make(Nodes),
		Usage:    make(map[ItemID][]NodeID),
		Variants: make(map[string]Items),
		Changes:  make(map[string]uuid.UUID),

		QualifierIndex:  make(map[string]ItemID),
		PartNumberIndex: make(map[string]ItemID),
		PathIndex:       make(map[string]NodeID),
	}
}

func (g *BOMGraph) MergeGraph(g2 *BOMGraph) error {
	if g2 == nil {
		return nil
	}
	maps.Copy(g.Items, g2.Items)
	maps.Copy(g.Changes, g2.Changes)
	maps.Copy(g.Variants, g2.Variants)
	maps.Copy(g.QualifierIndex, g2.QualifierIndex)
	maps.Copy(g.PartNumberIndex, g2.PartNumberIndex)
	maps.Copy(g.PathIndex, g2.PathIndex)
	return nil
}

func (g *BOMGraph) assignIDs() {
	for _, item := range g.Items {
		id, ok := g.QualifierIndex[item.Qualifier]
		if !ok {
			id = uuid.New()
			g.Changes[item.Qualifier] = id
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

func (g *BOMGraph) RootNode() *model.ItemNode {
	return g.Nodes[g.Root]
}

func (g *BOMGraph) RootItem() *model.Item {
	return g.Items[g.RootNode().ItemID]
}

func (g *BOMGraph) AddItem(item *model.Item) error {
	if item == nil {
		return errors.New("nil item")
	}
	_, exist := g.QualifierIndex[item.Qualifier]
	if exist {
		return errors.New("existed item")
	}
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	g.Items[item.ID] = item
	return nil
}

func (g *BOMGraph) AddNode(node *model.ItemNode) error {
	g.Nodes[node.ID] = node
	g.Usage[node.ItemID] = append(g.Usage[node.ItemID], node.ID)
	return nil
}

func (g *BOMGraph) refItemIDMapping(ref *BOMGraph) map[ItemID]ItemID {
	qualifierMapping := make(map[string]string)
	matched := make(map[string]bool)
	notMatched := make([][]string, 0)
	toMatch := make([][]string, 0)
	for _, item := range g.Items {
		refItemID, ok := ref.QualifierIndex[item.Qualifier]
		if ok {
			refItem := ref.Items[refItemID]
			qualifierMapping[item.Qualifier] = refItem.Qualifier
			matched[refItem.Qualifier] = true
		} else {
			toMatch = append(toMatch, strings.Split(item.Qualifier, "."))
		}
	}
	for _, item := range ref.Items {
		_, ok := matched[item.Qualifier]
		if !ok {
			notMatched = append(notMatched, strings.Split(item.Qualifier, "."))
		}
	}
	greedy := match.GreedyMatch(toMatch, notMatched)
	maps.Copy(qualifierMapping, greedy)

	itemIDMapping := make(map[ItemID]ItemID)
	for _, item := range g.Items {
		refItemQualifier, ok := qualifierMapping[item.Qualifier]
		if ok {
			itemIDMapping[item.ID] = ref.QualifierIndex[refItemQualifier]
		}
	}
	return itemIDMapping
}

func (g *BOMGraph) refNodeIDMapping(ref *BOMGraph) map[NodeID]NodeID {
	pathMapping := make(map[string]string)
	matched := make(map[string]bool)
	notMatched := make([][]string, 0)
	toMatch := make([][]string, 0)
	for _, node := range g.Nodes {
		refNodeID, ok := ref.PathIndex[node.Path]
		if ok {
			refNode := ref.Nodes[refNodeID]
			pathMapping[node.Path] = refNode.Path
			matched[refNode.Path] = true
		} else {
			toMatch = append(toMatch, strings.Split(node.Path, "/"))
		}
	}
	for _, node := range ref.Nodes {
		_, ok := matched[node.Path]
		if !ok {
			notMatched = append(notMatched, strings.Split(node.Path, "/"))
		}
	}
	greedy := match.GreedyMatch(toMatch, notMatched)
	maps.Copy(pathMapping, greedy)

	nodeIDMapping := make(map[NodeID]NodeID)
	for _, node := range g.Nodes {
		refNodePath, ok := pathMapping[node.Path]
		if ok {
			nodeIDMapping[node.ID] = ref.PathIndex[refNodePath]
		}
	}
	return nodeIDMapping
}

func (g *BOMGraph) Reference(ref *BOMGraph) *BOMGraph {
	ret := NewBOMGraph()
	itemIDMapping := g.refItemIDMapping(ref)
	for _, item := range g.Items {
		newItem := *item
		refItemID, ok := itemIDMapping[item.ID]
		if ok {
			newItem.ID = refItemID
			newItem.PartNumber = ref.Items[refItemID].PartNumber
		}
		ret.AddItem(&newItem)
	}
	nodeIDMapping := g.refNodeIDMapping(ref)
	for _, node := range g.Nodes {
		newNode := *node
		refNodeID, ok := nodeIDMapping[node.ID]
		if ok {
			newNode.ID = refNodeID
		}
		refItemID, ok := itemIDMapping[node.ItemID]
		if ok {
			newNode.ItemID = refItemID
		}
		refParentID, ok := nodeIDMapping[node.ParentID]
		if ok {
			newNode.ParentID = refParentID
		}
		newNode.Children = make([]NodeID, len(node.Children))
		for i, childID := range node.Children {
			refChildID, ok := nodeIDMapping[childID]
			if !ok {
				refChildID = childID
			}
			newNode.Children[i] = refChildID
		}
		ret.AddNode(&newNode)
	}
	rootID := g.RootNode().ID
	refRoot, ok := nodeIDMapping[rootID]
	if !ok {
		refRoot = rootID
	}
	ret.Root = refRoot
	ret.BuildIndex()
	ret.SortUsage()
	return ret
}

func (g *BOMGraph) SortUsage() {
	for _, nodeIDs := range g.Usage {
		sort.Slice(nodeIDs, func(i, j int) bool {
			return nodeIDs[i].String() < nodeIDs[j].String()
		})
	}
}
