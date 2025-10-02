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
