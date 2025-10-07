package hcl

import (
	"errors"
	"log/slog"

	"github.com/tychonis/cyanotype/internal/bomtree"
	"github.com/tychonis/cyanotype/model"
)

func getImplicitProcessQualifier(item *model.Item) string {
	return item.Qualifier + ".__process__"
}

func getImplicitCoProcessQualifier(item *model.Item) string {
	return item.Qualifier + ".__coprocess__"
}

func getImplicitCoItemQualifier(item *model.Item) string {
	return item.Qualifier + ".__coitem__"
}

// func (c *Core) findImplicitProcess(item *model.Item) (*model.Process, error) {
// 	q := getImplicitProcessQualifier(item)
// 	sym, err := c.Catalog.Find(q)
// 	if err != nil {
// 		return nil, err
// 	}
// 	p, ok := sym.(*model.Process)
// 	if !ok {
// 		return nil, errors.New("incorrect type for implicit process")
// 	}
// 	return p, nil
// }

func (c *Core) findProcesses(item *model.Item) ([]*model.Process, error) {
	return c.Catalog.GetItemProcesses(item.Digest)
}

func (c *Core) findCoProcesses(item *model.CoItem) ([]*model.CoProcess, error) {
	return c.Catalog.GetItemCoProcesses(item.Digest)
}

func (c *Core) build(coitem *model.CoItem, qty float64) (*bomtree.Node, error) {
	node := &bomtree.Node{
		CoItem:   coitem,
		Children: make([]*bomtree.Node, 0),
		Qty:      qty,
	}
	cp, err := c.findCoProcesses(coitem)
	if err != nil {
		return nil, err
	}
	if len(cp) <= 0 {
		slog.Info("No coprocess producing coitem", "qualifier", coitem.GetQualifier())
	}
	if len(cp) > 1 {
		slog.Warn("Multiple coprocesses can produce coitem", "qualifier", coitem.GetQualifier())
	}
	coProcess := cp[0]
	node.CoProcess = coProcess

	itemID := coProcess.Input[0].Item
	itemSym, err := c.Catalog.Get(itemID)
	if err != nil {
		return nil, err
	}
	item, ok := itemSym.(*model.Item)
	if !ok {
		return nil, errors.New("invalid input")
	}
	node.Item = item

	p, err := c.findProcesses(item)
	if err != nil {
		return nil, err
	}
	if len(p) <= 0 {
		slog.Info("No process producing item", "qualifier", item.GetQualifier())
		return node, nil
	}
	if len(p) > 1 {
		slog.Warn("Multiple processes can produce item", "qualifier", item.GetQualifier())
	}
	process := p[0]

	node.Process = process
	for _, input := range process.Input {
		child, err := c.Catalog.Get(input.Item)
		if err != nil {
			return nil, err
		}
		childCoItem, ok := child.(*model.CoItem)
		if !ok {
			return nil, errors.New("invalid input")
		}
		childNode, err := c.build(childCoItem, input.Qty)
		if err != nil {
			return nil, err
		}
		childNode.Parent = node
		node.Children = append(node.Children, childNode)
	}
	return node, nil
}

func (c *Core) Build(item *model.Item) (*bomtree.Node, error) {
	coItems, err := c.Catalog.GetCoItems(item.Digest)
	if err != nil {
		return nil, err
	}
	if len(coItems) != 1 {
		return nil, errors.New("multiple coitems not implemented yet")
	}
	coItemSym, err := c.Catalog.Get(coItems[0].Item)
	if err != nil {
		return nil, err
	}
	coItem, ok := coItemSym.(*model.CoItem)
	if !ok {
		return nil, errors.New("not a coitem")
	}
	return c.build(coItem, 1)
}
