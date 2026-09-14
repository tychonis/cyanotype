package scene

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/tychonis/cyanotype/core/bomtree"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/instantiator"
	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/core/ranker"
	"github.com/tychonis/cyanotype/model"
)

func getItemModel(item *model.Item) *model.Artifact {
	for _, artifact := range item.Content.Artifacts {
		if artifact.Tag == "model" {
			return artifact
		}
	}
	return nil
}

func renderLeafNode(node *bomtree.Node) (*model.Scene, error) {
	m := getItemModel(node.Item)
	if m == nil {
		return &model.Scene{
			Objects: nil,
		}, nil
	}
	comps := strings.Split(m.Filename, ".")
	obj := model.Object{
		Name:      node.Name,
		Item:      node.Item.Digest,
		ItemName:  node.Item.GetName(),
		Source:    m.Source,
		Type:      "." + comps[len(comps)-1],
		Placement: &model.IdentityPlacement,
	}
	return &model.Scene{
		Objects: []*model.Object{&obj},
	}, nil
}

func renderCacheNode(node *bomtree.Node) (*model.Scene, error) {
	obj := model.Object{
		Name:      node.Name,
		Item:      node.Item.Digest,
		ItemName:  node.Item.GetName(),
		Source:    node.Item.Digest,
		Type:      "cache",
		Placement: &model.IdentityPlacement,
	}
	return &model.Scene{
		Objects: []*model.Object{&obj},
	}, nil
}

func buildPlacementMap(node *bomtree.Node) map[string]model.Placement {
	placement := make(map[string]model.Placement)
	p := node.Process
	if p.Content.GetType() == process.DRAWING {
		content, ok := p.Content.(*process.Drawing)
		if !ok {
			return nil
		}
		for _, comp := range content.Components {
			placement[comp.Name] = model.Placement{
				Rotation: *comp.Rotation,
				Position: *comp.Translation,
			}
		}
	}
	return placement
}

func renderNode(node *bomtree.Node, recursive bool) (*model.Scene, error) {
	slog.Debug("rendering node", "node", node.Name)
	if len(node.Children) <= 0 {
		return renderLeafNode(node)
	}

	placement := buildPlacementMap(node)

	objs := []*model.Object{}
	for _, child := range node.Children {
		var childScene *model.Scene
		var err error
		if recursive {
			childScene, err = renderNode(child, recursive)
		} else {
			childScene, err = renderCacheNode(child)
		}
		if err != nil {
			return nil, err
		}
		for _, obj := range childScene.Objects {
			childPlacement, ok := placement[child.Name]
			if !ok {
				childPlacement = model.IdentityPlacement
			}
			obj.Placement = obj.Placement.Compose(&childPlacement)
		}
		objs = append(objs, childScene.Objects...)
	}
	return &model.Scene{
		Objects: objs,
	}, nil
}

func RenderSymbol(cat *catalog.Catalog, sym model.ConcreteSymbol, recursive bool) (*model.Scene, error) {
	ins := instantiator.New()
	ins.Ranker = ranker.NewCatalogTypeRanker(cat, process.DRAWING)
	var coItem *model.CoItem
	var ok bool
	var err error
	switch parent := sym.(type) {
	case *model.CoItem:
		coItem = parent
	case *model.Item:
		coItemSym, err := cat.FindCurrent(qualifier.ImplicitCoItem(parent))
		if err != nil {
			return nil, err
		}
		coItem, ok = coItemSym.(*model.CoItem)
		if !ok {
			return nil, fmt.Errorf("failed to assert symbol as CoItem")
		}
	default:
		return nil, fmt.Errorf("unsupported symbol type")
	}

	var node *bomtree.Node
	if recursive {
		node, err = ins.InstantiateTree(cat, "tmp", coItem)
	} else {
		node, err = ins.ExpandNode(cat, "tmp", coItem)
	}
	if err != nil {
		return nil, err
	}
	return renderNode(node, recursive)
}
