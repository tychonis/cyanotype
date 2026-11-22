package instantiator

import (
	"errors"

	"github.com/tychonis/cyanotype/core/bomtree"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/core/ranker"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

type Instantiator struct {
	Catalog *catalog.Catalog
	Ranker  ranker.Ranker
}

func New(cat *catalog.Catalog, r ranker.Ranker) *Instantiator {
	if r == nil {
		r = &ranker.NaiveRanker{}
	}
	return &Instantiator{
		Catalog: cat,
		Ranker:  r,
	}
}

func (i *Instantiator) instantiateNode(name string, coitem *model.CoItem, qty float64) (*bomtree.Node, error) {
	node := &bomtree.Node{
		Name:     name,
		CoItem:   coitem,
		Children: make([]*bomtree.Node, 0),
		Qty:      qty,
	}
	cp, err := i.Catalog.GetItemCoProcesses(coitem.Digest)
	if err != nil {
		return nil, err
	}
	coProcess, err := i.Ranker.TopCoProcess(cp)
	if err != nil {
		return nil, err
	}
	node.CoProcess = coProcess

	itemID := coProcess.Input[0].Item
	itemSym, err := i.Catalog.Get(itemID)
	if err != nil {
		return nil, err
	}
	item, ok := itemSym.(*model.Item)
	if !ok {
		return nil, errors.New("invalid input")
	}
	node.Item = item

	p, err := i.Catalog.GetItemProcesses(item.Digest)
	if err != nil {
		return nil, err
	}
	process, err := i.Ranker.TopProcess(p)
	if err != nil {
		return nil, err
	}

	node.Process = process
	return node, nil
}

func (i *Instantiator) instantiate(name string, coitem *model.CoItem, qty float64) (*bomtree.Node, error) {
	node, err := i.instantiateNode(name, coitem, qty)
	if err != nil {
		return nil, err
	}
	for _, input := range node.Process.Input {
		child, err := i.Catalog.Get(input.Item)
		if err != nil {
			return nil, err
		}
		childCoItem, ok := child.(*model.CoItem)
		if !ok {
			return nil, errors.New("invalid input")
		}
		childNode, err := i.instantiate(input.Name, childCoItem, input.Qty)
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

func (i *Instantiator) InstantiateTree(name string, coItem *model.CoItem) (*bomtree.Node, error) {
	return i.instantiate(name, coItem, 1)
}

// InstantiateTreeFromItem provides a shortcut. Trees should be instantiated from a coitem.
func (i *Instantiator) InstantiateTreeFromItem(name string, item *model.Item) (*bomtree.Node, error) {
	coItemQualifier := qualifier.ImplicitCoItem(item)
	coItemSym, err := i.Catalog.FindCurrent(coItemQualifier)
	if err != nil {
		return nil, err
	}
	coItem, ok := coItemSym.(*model.CoItem)
	if !ok {
		return nil, errors.New("not a coitem")
	}
	return i.InstantiateTree(name, coItem)
}

func (i *Instantiator) ExpandNode(name string, coitem *model.CoItem) (*bomtree.Node, error) {
	node, err := i.instantiateNode(name, coitem, 1)
	if err != nil {
		return nil, err
	}
	for _, input := range node.Process.Input {
		child, err := i.Catalog.Get(input.Item)
		if err != nil {
			return nil, err
		}
		childCoItem, ok := child.(*model.CoItem)
		if !ok {
			return nil, errors.New("invalid input")
		}
		childNode, err := i.instantiateNode(input.Name, childCoItem, input.Qty)
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

func (i *Instantiator) TreeFromQualifier(root string) (*bomtree.Node, error) {
	sym, err := i.Catalog.FindCurrent(root)
	if err != nil {
		return nil, err
	}

	switch resolved := sym.(type) {
	case *model.CoItem:
		return i.InstantiateTree(resolved.GetName(), resolved)
	case *model.Item:
		// TODO: Deprecate this.
		return i.InstantiateTreeFromItem(resolved.GetName(), resolved)
	default:
		return nil, errors.New("unknown symbol type")
	}
}
