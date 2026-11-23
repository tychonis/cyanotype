package scene

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/tychonis/cyanotype/core/bomtree"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/instantiator"
	"github.com/tychonis/cyanotype/core/operation"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/core/ranker"
	"github.com/tychonis/cyanotype/model"
)

func getItemModel(item *model.Item) *model.Artifact {
	for _, artifact := range item.Artifacts {
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
		Placement: model.IdentityPlacement,
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
		Placement: model.IdentityPlacement,
	}
	return &model.Scene{
		Objects: []*model.Object{&obj},
	}, nil
}

// hackyway to get item and process from a qualifier point
// returns item, process, error
func getItemAndOperationFromPoint(point model.Qualifier) (model.Qualifier, model.Qualifier, error) {
	components := strings.Split(string(point), ".")
	if len(components) < 2 {
		return "", "", fmt.Errorf("invalid qualifier format")
	}
	item := strings.Join(components[:len(components)-1], ".")
	fakeItem := &model.Item{}
	fakeItem.Qualifier = model.Qualifier(item)
	process := qualifier.ImplicitProcess(fakeItem)
	operation := process + "." + components[len(components)-1]
	return model.Qualifier(item), model.Qualifier(operation), nil
}

func getAlignPoint(ins *instantiator.Instantiator, point model.Qualifier) (*AlignPoint, error) {
	// pointSymbol, err := ins.Catalog.Get(point)
	// if err != nil {
	// 	return nil, err
	// }
	// contract, ok := pointSymbol.(*model.Contract)
	// if !ok {
	// 	return nil, fmt.Errorf("symbol is not a contract")
	// }
	// return &AlignPoint{
	// 	Child: child,
	// 	Frame: placement,
	// }, nil
	return nil, nil
}

func solveConstraints(ins *instantiator.Instantiator, constraints []*model.Operation, placements map[string]model.Placement) error {
	alignConstraints := make([]AlignConstraint, 0)
	for _, op := range constraints {
		points, err := operation.AlignFromOperation(op)
		if err != nil {
			return err
		}
		alignPoints := make([]AlignPoint, 0, len(points))
		for _, point := range points {
			ap, err := getAlignPoint(ins, point)
			if err != nil {
				return err
			}
			alignPoints = append(alignPoints, *ap)
		}
		for _, p := range alignPoints[1:] {
			alignConstraints = append(alignConstraints, AlignConstraint{
				A: alignPoints[0],
				B: p,
			})
		}
	}
	return solveAlignConstraints(alignConstraints, placements)
}

func buildPlacementMap(ins *instantiator.Instantiator, node *bomtree.Node) (map[string]model.Placement, error) {
	placements := make(map[string]model.Placement)
	p := node.Process
	constraints := make([]*model.Operation, 0)

	for _, opID := range p.Operations {
		sym, err := ins.Catalog.Get(opID)
		if err != nil {
			continue
		}
		op, ok := sym.(*model.Operation)
		if !ok {
			continue
		}
		if op.Category == "placement" {
			child, placement, err := operation.PlacementFromOperation(op)
			if err != nil {
				continue
			}
			placements[child] = placement
		}
		if op.Category == "align" {
			constraints = append(constraints, op)
		}
	}
	err := solveConstraints(ins, constraints, placements)
	return placements, err
}

func renderNode(ins *instantiator.Instantiator, node *bomtree.Node, recursive bool) (*model.Scene, error) {
	slog.Debug("rendering node", "node", node.Name)
	if len(node.Children) <= 0 {
		if recursive {
			return renderLeafNode(node)
		} else {
			return renderCacheNode(node)
		}
	}

	placement, err := buildPlacementMap(ins, node)
	if err != nil {
		return nil, err
	}

	objs := []*model.Object{}
	for _, child := range node.Children {
		var childScene *model.Scene
		var err error
		if recursive {
			childScene, err = renderNode(ins, child, recursive)
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
			obj.Placement = obj.Placement.Compose(childPlacement)
		}
		objs = append(objs, childScene.Objects...)
	}
	return &model.Scene{
		Objects: objs,
	}, nil
}

func RenderSymbol(cat *catalog.Catalog, sym model.ConcreteSymbol, recursive bool) (*model.Scene, error) {
	r := ranker.NewCatalogTypeRanker(cat, "drawing")
	ins := instantiator.New(cat, r)
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
		return nil, fmt.Errorf("unsupported symbol type for rendering")
	}

	var node *bomtree.Node
	if recursive {
		node, err = ins.InstantiateTree("tmp", coItem)
	} else {
		node, err = ins.ExpandNode("tmp", coItem)
	}
	if err != nil {
		return nil, err
	}
	return renderNode(ins, node, recursive)
}
