package bomtree

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model"
)

type Node struct {
	ID   model.Digest
	Name string

	CoItem    *model.CoItem
	CoProcess *model.CoProcess
	Item      *model.Item
	Process   *model.Process

	Path     string
	Parent   *Node
	Children []*Node
	Qty      float64
}

func (node *Node) Count() map[string]float64 {
	counter := make(map[string]float64)
	count(node, 1, counter)
	return counter
}

func count(node *Node, multiplier float64, counter map[string]float64) {
	item := node.Item

	if len(node.Children) == 0 {
		counter[item.Qualifier] += multiplier
		return
	}

	// also count assembly?
	counter[item.Qualifier] += multiplier

	for _, child := range node.Children {
		count(child, child.Qty*multiplier, counter)
	}
}

type NodeInfo struct {
	ID       model.Digest   `json:"id"`
	Name     string         `json:"name"`
	Item     model.Digest   `json:"item"`
	Parent   model.Digest   `json:"parent"`
	Children []model.Digest `json:"children"`
	Qty      float64        `json:"qty"`
}

type TreeDocument struct {
	Root  model.Digest                    `json:"root"`
	Nodes map[model.Digest]*NodeInfo      `json:"nodes"`
	Reuse map[model.Digest][]model.Digest `json:"reuse"`
}

func (node *Node) Export() ([]byte, error) {
	doc := &TreeDocument{
		Root:  node.ID,
		Nodes: make(map[model.Digest]*NodeInfo),
		Reuse: make(map[model.Digest][]model.Digest),
	}
	export(node, doc)
	return json.Marshal(doc)
}

func export(node *Node, doc *TreeDocument) {
	parentID := ""
	if node.Parent != nil {
		doc.Nodes[node.Parent.ID].Children = append(doc.Nodes[node.Parent.ID].Children, node.ID)
		parentID = node.Parent.ID
	}

	info := &NodeInfo{
		ID:       node.ID,
		Name:     node.Name,
		Item:     node.Item.Digest,
		Parent:   parentID,
		Children: make([]model.Digest, 0, len(node.Children)),
		Qty:      node.Qty,
	}
	doc.Nodes[node.ID] = info

	reuse, ok := doc.Reuse[node.Item.Digest]
	if !ok {
		reuse = make([]model.Digest, 0)
	}
	doc.Reuse[node.Item.Digest] = append(reuse, node.ID)

	for _, child := range node.Children {
		export(child, doc)
	}
}
