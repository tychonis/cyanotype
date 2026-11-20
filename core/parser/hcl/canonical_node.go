package hcl

import (
	"log/slog"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/model"
)

type SymbolContext[T model.ConcreteSymbol] struct {
	Symbol T
	Blocks []*hclsyntax.Block
}

type CanonicalNode struct {
	ExplicitBlock *hclsyntax.Block
	CoItem        *SymbolContext[*model.CoItem]
	CoProcess     *SymbolContext[*model.CoProcess]
	Item          *SymbolContext[*model.Item]
	Process       *SymbolContext[*model.Process]
}

func (p *Parser) buildCanonicalNode(ctx *ParserContext, item *model.Item) (*CanonicalNode, error) {
	slog.Debug("building canonical node", "module", ctx.CurrentModule(), "item", item.Qualifier)

	canonicalNode := &CanonicalNode{}

	canonicalNode.Item = &SymbolContext[*model.Item]{
		Symbol: item,
		Blocks: make([]*hclsyntax.Block, 0),
	}

	co := &model.CoItem{}
	co.Qualifier = qualifier.ImplicitCoItem(item)
	co.Name = "Implicit requirements for " + item.Name
	canonicalNode.CoItem = &SymbolContext[*model.CoItem]{
		Symbol: co,
		Blocks: make([]*hclsyntax.Block, 0),
	}

	cp := &model.CoProcess{}
	cp.Name = "Implicit coprocess for " + item.Name
	cp.Qualifier = qualifier.ImplicitCoProcess(item)
	canonicalNode.CoProcess = &SymbolContext[*model.CoProcess]{
		Symbol: cp,
		Blocks: make([]*hclsyntax.Block, 0),
	}

	proc := &model.Process{}
	proc.Name = "Implicit process for " + item.Name
	proc.Qualifier = qualifier.ImplicitProcess(item)
	canonicalNode.Process = &SymbolContext[*model.Process]{
		Symbol: proc,
		Blocks: make([]*hclsyntax.Block, 0),
	}

	return canonicalNode, nil
}

// AssignSubBlocks assigns the sub-blocks of a given HCL block to the appropriate sections of the canonical node.
// Current assumption is that this is only applied for item blocks as syntax sugar.
func (p *Parser) AssignSubBlocks(ctx *ParserContext, node *CanonicalNode, block *hclsyntax.Block) error {
	node.ExplicitBlock = block
	for _, child := range block.Body.Blocks {
		switch child.Type {
		case "artifact":
			node.Item.Blocks = append(node.Item.Blocks, child)
		case "contract":
			node.Item.Blocks = append(node.Item.Blocks, child)
			node.Process.Blocks = append(node.Process.Blocks, child)
			node.CoItem.Blocks = append(node.CoItem.Blocks, child)
		case "operation":
			node.Process.Blocks = append(node.Process.Blocks, child)
		case "component":
			node.Process.Blocks = append(node.Process.Blocks, child)
		default:
			slog.Warn("Unrecognized sub block, skipping", slog.String("type", child.Type))
			continue
		}
	}
	return nil
}

func (p *Parser) processItemArtifactBlock(ctx *ParserContext, item *model.Item, block *hclsyntax.Block) error {
	artifact, err := p.processArtifactBlock(ctx, block)
	if err != nil {
		return err
	}
	if item.Artifacts == nil {
		item.Artifacts = make([]*model.Artifact, 0)
	}
	item.Artifacts = append(item.Artifacts, artifact)
	return nil
}

func (p *Parser) processItemContractBlock(ctx *ParserContext, item *model.Item, block *hclsyntax.Block) error {
	ctx.CurrentSymbol = item
	defer func() { ctx.CurrentSymbol = nil }()
	contract, err := p.processContractBlock(ctx, block)
	if err != nil {
		return err
	}
	if item.Implement == nil {
		item.Implement = make([]model.ContractID, 0)
	}
	item.Implement = append(item.Implement, contract.Digest)
	return nil
}

func (p *Parser) processItemSubBlocks(ctx *ParserContext, symCtx *SymbolContext[*model.Item]) error {
	for _, block := range symCtx.Blocks {
		switch block.Type {
		case "artifact":
			err := p.processItemArtifactBlock(ctx, symCtx.Symbol, block)
			if err != nil {
				return err
			}
		case "contract":
			err := p.processItemContractBlock(ctx, symCtx.Symbol, block)
			if err != nil {
				return err
			}
		default:
			slog.Warn("Unsupported block for item, skipping", slog.String("type", block.Type))
			continue
		}
	}
	return nil
}

func (p *Parser) processCoItemSubBlocks(ctx *ParserContext, symCtx *SymbolContext[*model.CoItem]) error {
	for _, block := range symCtx.Blocks {
		switch block.Type {
		case "contract":
			continue
		default:
			slog.Warn("Unsupported block for co-item, skipping", slog.String("type", block.Type))
			continue
		}
	}
	return nil
}

func (p *Parser) processCoProcessSubBlocks(ctx *ParserContext, symCtx *SymbolContext[*model.CoProcess]) error {
	for _, block := range symCtx.Blocks {
		switch block.Type {
		default:
			slog.Warn("Unsupported block for co-process, skipping", slog.String("type", block.Type))
			continue
		}
	}
	return nil
}

func (p *Parser) processCompanionProcess(ctx *ParserContext, node *CanonicalNode) error {
	components, err := p.extractUnresolvedBOMLines(ctx, node.ExplicitBlock)
	if err != nil {
		return err
	}
	operations := make([]*model.Operation, 0)
	ctx.CurrentSymbol = node.Process.Symbol
	defer func() { ctx.CurrentSymbol = nil }()
	for _, block := range node.Process.Blocks {
		switch block.Type {
		case "contract":
			operation, err := p.generateImplicitOperation(ctx, block)
			if err != nil {
				return err
			}
			if operation != nil {
				operations = append(operations, operation)
			}
		case "operation":
			operation, err := p.processOperationBlock(ctx, block)
			if err != nil {
				return err
			}
			if operation != nil {
				operations = append(operations, operation)
			}
		case "component":
			comp, err := p.processComponentBlock(ctx, block)
			if err != nil {
				return err
			}
			components = append(components, comp)
		default:
			slog.Warn("Unsupported block for process, skipping", slog.String("type", block.Type))
		}
	}
	ops, err := p.resolveBOMLines(ctx, node.Process.Symbol, components)
	if err != nil {
		return err
	}
	node.Process.Symbol.Output = []*model.BOMLine{
		{
			Item: node.Item.Symbol.Digest,
			Qty:  1,
		},
	}
	operations = append(operations, ops...)
	node.Process.Symbol.Operations = make([]model.OperationID, 0, len(operations))
	for _, op := range operations {
		slog.Debug("fusing operation", "operation", op)
		err := p.Fuse(op)
		if err != nil {
			return err
		}
		node.Process.Symbol.Operations = append(node.Process.Symbol.Operations, op.Digest)
	}

	return nil
}

// extractUnresolvedBOMLines extracts the unresolved BOM lines from the canonical node using the "from" attribute if present.
// If the "from" attribute is not present, it returns an abstract process content.
// TODO: Deprecate this function in favor of sub block definition "component"
func (p *Parser) extractUnresolvedBOMLines(ctx *ParserContext, explicitBlock *hclsyntax.Block) ([]*UnresolvedBOMLine, error) {
	attrs, err := extractAttributes(explicitBlock.Body)
	if err != nil {
		return nil, err
	}
	var from []*UnresolvedBOMLine
	fromAttr, ok := attrs["from"]
	if ok {
		from = parseBOMLinesAttr(ctx, fromAttr)
	}
	return from, nil
}

func (p *Parser) ProcessCanonicalNode(ctx *ParserContext, node *CanonicalNode) error {
	err := p.processItemSubBlocks(ctx, node.Item)
	if err != nil {
		return err
	}

	err = p.Fuse(node.Item.Symbol)
	if err != nil {
		return err
	}

	err = p.processCoItemSubBlocks(ctx, node.CoItem)
	if err != nil {
		return err
	}

	err = p.Fuse(node.CoItem.Symbol)
	if err != nil {
		return err
	}

	err = p.processCoProcessSubBlocks(ctx, node.CoProcess)
	if err != nil {
		return err
	}

	node.CoProcess.Symbol.Input = []*model.BOMLine{
		{
			Item: node.Item.Symbol.Digest,
			Qty:  1,
		},
	}
	node.CoProcess.Symbol.Output = []*model.BOMLine{
		{
			Item: node.CoItem.Symbol.Digest,
			Qty:  1,
		},
	}

	err = p.Fuse(node.CoProcess.Symbol)
	if err != nil {
		return err
	}

	err = p.processCompanionProcess(ctx, node)
	if err != nil {
		return err
	}

	err = p.Fuse(node.Process.Symbol)
	if err != nil {
		return err
	}

	return nil
}
