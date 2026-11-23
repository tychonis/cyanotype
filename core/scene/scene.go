package scene

import (
	"errors"
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

type SymbolAttribute struct {
	Symbol        model.ConcreteSymbol
	RemainingPath string
}

// TODO: Find out a better way to resolve symbol attributes from a qualifier point
func resolveSymbolAttribute(ins *instantiator.Instantiator, point model.Qualifier) (*SymbolAttribute, error) {
	qualifier := string(point)

	for {
		sym, err := ins.Catalog.FindCurrent(qualifier)
		if err == nil {
			suffix := strings.TrimPrefix(point, qualifier)
			suffix = strings.TrimPrefix(suffix, ".")

			return &SymbolAttribute{
				Symbol:        sym,
				RemainingPath: suffix,
			}, nil
		}
		if !errors.Is(err, catalog.ErrNotFound) {
			return nil, err
		}

		i := strings.LastIndex(qualifier, ".")
		if i <= 0 {
			return nil, fmt.Errorf("cannot resolve symbol attribute %q: %w",
				point, catalog.ErrNotFound)
		}
		qualifier = qualifier[:i]
	}
}

func getOperation(ins *instantiator.Instantiator, node *bomtree.Node, name string) (*model.Operation, error) {
	for _, opID := range node.Process.Operations {
		sym, err := ins.Catalog.Get(opID)
		if err != nil {
			continue
		}
		op, ok := sym.(*model.Operation)
		if !ok {
			continue
		}
		if op.Name == name {
			return op, nil
		}
		slog.Info("looking for operation", "opID", opID, "name", op.Name)

	}
	return nil, fmt.Errorf("operation %s not found in node %s", name, node.Item.GetName())
}

func getAlignPoint(ins *instantiator.Instantiator, point model.Qualifier) (*AlignPoint, error) {
	pAttr, err := resolveSymbolAttribute(ins, point)
	if err != nil {
		return nil, err
	}
	child := strings.Split(pAttr.RemainingPath, ".")
	proc, ok := pAttr.Symbol.(*model.Process)
	if !ok {
		slog.Info("symbol is not a process", "symbol", pAttr.Symbol, "remainingPath", pAttr.RemainingPath)
		return nil, fmt.Errorf("symbol is not a process")
	}

	if len(child) < 2 {
		return nil, fmt.Errorf("invalid attribute path %s for process %s", pAttr.RemainingPath, proc.Name)
	}

	bomLine := proc.GetBomLine(child[0])
	if bomLine == nil {
		return nil, fmt.Errorf("cannot find BOM line %s in process %s", child[0], proc.Name)
	}
	coitemSym, err := ins.Catalog.Get(bomLine.Item)
	if err != nil {
		return nil, err
	}
	coitem, ok := coitemSym.(*model.CoItem)
	if !ok {
		return nil, fmt.Errorf("symbol is not an coitem")
	}
	node, err := ins.ExpandNode("tmp", coitem)
	if err != nil {
		return nil, err
	}
	op, err := getOperation(ins, node, child[1])
	if err != nil {
		return nil, err
	}
	placement, err := operation.PointFromOperation(op)
	if err != nil {
		return nil, err
	}

	return &AlignPoint{
		Child: child[0],
		Frame: placement,
	}, nil
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
	slog.Info("placement map final", "placement", placement)

	objs := []*model.Object{}
	for _, child := range node.Children {
		slog.Info("rendering child", "child", child.Name)
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
