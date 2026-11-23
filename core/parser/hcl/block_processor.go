package hcl

import (
	"errors"
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/cerror"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) Fuse(sym model.ConcreteSymbol) error {
	if sym.GetDigest() != "" {
		slog.Debug("symbol already fused", "qualifier", sym.GetQualifier())
		return errors.New("symbol already fused")
	}
	dig, err := digest.SHA256FromSymbol(sym)
	if err != nil {
		return err
	}
	sym.SetDigest(dig)
	return p.Symbols.RegisterConcreteSymbol(sym)
}

func (p *Parser) ProcessSymbol(s *UnprocessedSymbol) (sym model.ConcreteSymbol, err error) {
	// Already processed before.
	if s.qualifier != "" {
		return p.Symbols.FindConcreteSymbol(s.qualifier)
	}

	// New symbol.
	if s.Block == nil || s.Context == nil {
		return nil, errors.New("illegal nil symbol")
	}
	switch s.Block.Type {
	case "item":
		sym, err = p.processItemBlock(s.Context, s.Block)
	case "coitem":
		sym, err = p.processCoItemBlock(s.Context, s.Block)
	case "process":
		sym, err = p.processProcessBlock(s.Context, s.Block)
	case "coprocess":
		sym, err = p.processCoProcessBlock(s.Context, s.Block)
	case "contract":
		sym, err = p.processContractBlock(s.Context, s.Block)
	default:
		return nil, cerror.ErrorWithRange(errors.New("unknown block type"), s.Block.Range())
	}
	if err == nil {
		s.qualifier = sym.GetQualifier()
	} else {
		err = cerror.ErrorWithRange(err, s.Block.Range())
	}
	return
}

var RESERVED = map[string]struct{}{
	"part_number": {},
	"from":        {},
}

func (p *Parser) getDetails(ctx *ParserContext, attrs hcl.Attributes) (stable.Map, error) {
	keys := make([]string, 0)
	for key := range attrs {
		_, ok := RESERVED[key]
		if !ok {
			keys = append(keys, key)
		}
	}
	ret := make(stable.Map)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val, err := getString(attrs, key)
		if err != nil {
			return ret, err
		}
		ret[key] = val
	}
	return ret, nil
}

func (p *Parser) processItemBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.Item, error) {
	name := block.Labels[0]
	attrs, err := extractAttributes(block.Body)
	if err != nil {
		return nil, err
	}

	item := &model.Item{}
	item.Qualifier = ctx.NameToQualifier(name)
	item.Name = name
	item.Details, err = p.getDetails(ctx, attrs)
	if err != nil {
		return item, err
	}

	canonicalNode, err := p.buildCanonicalNode(ctx, item)
	if err != nil {
		return nil, err
	}

	err = p.AssignSubBlocks(ctx, canonicalNode, block)
	if err != nil {
		return nil, err
	}

	err = p.ProcessCanonicalNode(ctx, canonicalNode)
	if err != nil {
		return nil, err
	}

	return item, err
}

func (p *Parser) processCoItemBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.CoItem, error) {
	name := block.Labels[0]
	attrs, err := extractAttributes(block.Body)
	if err != nil {
		return nil, err
	}

	coItem := &model.CoItem{}
	coItem.Qualifier = ctx.NameToQualifier(name)
	coItem.Name = name

	coItem.Details, err = p.getDetails(ctx, attrs)
	if err != nil {
		return coItem, err
	}

	subContent, err := p.ProcessSubBlocks(ctx, block)
	if err != nil {
		return coItem, err
	}
	coItem.Artifacts = subContent.Artifacts

	err = p.Fuse(coItem)
	return coItem, err
}

func (p *Parser) processProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (ret *model.Process, err error) {
	name := block.Labels[0]
	ret = &model.Process{}
	ret.Qualifier = ctx.NameToQualifier(name)
	_, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	err = p.Fuse(ret)
	return
}

func (p *Parser) processCoProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (ret *model.CoProcess, err error) {
	name := block.Labels[0]
	ret = &model.CoProcess{}
	ret.Qualifier = ctx.NameToQualifier(name)
	_, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	err = p.Fuse(ret)
	return
}
