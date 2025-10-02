package bomtree

import (
	"github.com/tychonis/cyanotype/model/v2"
)

type Node struct {
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
