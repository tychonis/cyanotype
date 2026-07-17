package instantiator

import (
	"errors"

	"github.com/tychonis/cyanotype/core/bomtree"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/ranker"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/internal/qualifier"
	"github.com/tychonis/cyanotype/model"
)

type Instantiator struct {
	Ranker ranker.Ranker
}

func New() *Instantiator {
	return &Instantiator{
		Ranker: &ranker.NaiveRanker{},
	}
}

func (i *Instantiator) instantiate(cat *catalog.Catalog, name string, coitem *model.CoItem, qty float64) (*bomtree.Node, error) {
	node := &bomtree.Node{
		Name:     name,
		CoItem:   coitem,
		Children: make([]*bomtree.Node, 0),
		Qty:      qty,
	}
	cp, err := cat.GetItemCoProcesses(coitem.Digest)
	if err != nil {
		return nil, err
	}
	coProcess, err := i.Ranker.TopCoProcess(cp)
	if err != nil {
		return nil, err
	}
	node.CoProcess = coProcess

	itemID := coProcess.Input()[0].Item
	itemSym, err := cat.Get(itemID)
	if err != nil {
		return nil, err
	}
	item, ok := itemSym.(*model.Item)
	if !ok {
		return nil, errors.New("invalid input")
	}
	node.Item = item

	p, err := cat.GetItemProcesses(item.Digest)
	if err != nil {
		return nil, err
	}
	process, err := i.Ranker.TopProcess(p)
	if err != nil {
		return nil, err
	}

	node.Process = process
	for _, input := range process.Input() {
		child, err := cat.Get(input.Item)
		if err != nil {
			return nil, err
		}
		childCoItem, ok := child.(*model.CoItem)
		if !ok {
			return nil, errors.New("invalid input")
		}
		childNode, err := i.instantiate(cat, input.Name, childCoItem, input.Qty)
		if err != nil {
			return nil, err
		}
		childNode.Parent = node
		node.Children = append(node.Children, childNode)
	}
	node.ID, err = digest.RandomSHA256()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (i *Instantiator) InstantiateTree(cat *catalog.Catalog, name string, coItem *model.CoItem) (*bomtree.Node, error) {
	return i.instantiate(cat, name, coItem, 1)
}

// InstantiateTreeFromItem provides a shortcut. Trees should be instantiated from a coitem.
func (i *Instantiator) InstantiateTreeFromItem(cat *catalog.Catalog, name string, item *model.Item) (*bomtree.Node, error) {
	coItemQualifier := qualifier.ImplicitCoItem(item)
	coItemSym, err := cat.FindCurrent(coItemQualifier)
	if err != nil {
		return nil, err
	}
	coItem, ok := coItemSym.(*model.CoItem)
	if !ok {
		return nil, errors.New("not a coitem")
	}
	return i.InstantiateTree(cat, name, coItem)
}
