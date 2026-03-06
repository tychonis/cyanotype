package hcl

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/internal/cerror"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

func (c *Core) ParseSymbol(s *UnprocessedSymbol) (sym model.ConcreteSymbol, err error) {
	// Already parsed before.
	if s.qualifier != "" {
		return c.Catalog.Find(s.qualifier)
	}

	// New symbol.
	if s.Block == nil || s.Context == nil {
		return nil, errors.New("illegal nil symbol")
	}
	switch s.Block.Type {
	case "item":
		sym, err = c.parseItemBlock(s.Context, s.Block)
	case "coitem":
		sym, err = c.parseCoItemBlock(s.Context, s.Block)
	case "process":
		sym, err = c.parseProcessBlock(s.Context, s.Block)
	case "coprocess":
		sym, err = c.parseCoProcessBlock(s.Context, s.Block)
	case "contract":
		sym, err = c.parseContractBlock(s.Context, s.Block)
	default:
		return nil, cerror.ErrorWithRange("unknown block type", s.Block.Range())
	}
	if err == nil {
		// TODO: remove side effects here.
		s.qualifier = sym.GetQualifier()
		slog.Debug("adding symbol",
			"qualifier", sym.GetQualifier(), "digest", sym.GetDigest())
		err = c.Catalog.Add(sym)
	} else {
		err = cerror.ErrorWithRange(err.Error(), s.Block.Range())
	}
	return
}

func refToQualifier(ctx *ParserContext, ref []string) string {
	if ctx.CurrentModule() == "." {
		return "." + strings.Join(ref, ".")
	} else {
		return strings.Join(append([]string{ctx.CurrentModule()}, ref...), ".")
	}
}

func (c *Core) resolveContractsLinesAttr(ctx *ParserContext, attr *hcl.Attribute) ([]model.ContractID, error) {
	refs, err := c.readContractLine(ctx, attr.Expr)
	if err != nil {
		return nil, err
	}
	return c.resolveContractsID(ctx, refs)
}

var RESERVED = map[string]struct{}{
	"part_number": struct{}{},
	"source":      struct{}{},
	"from":        struct{}{},
	"impl":        struct{}{},
}

func (c *Core) getDetails(ctx *ParserContext, attrs hcl.Attributes) (stable.Map, error) {
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

func (c *Core) parseItemBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.Item, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	pn, _ := getString(attrs, "part_number")
	// ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")

	var input []*model.BOMLine
	var err error
	fromAttr, ok := attrs["from"]
	if ok {
		from := parseBOMLinesAttr(ctx, fromAttr)
		input, err = c.processKeywordFROM(ctx, from)
		if err != nil {
			return nil, err
		}
	}

	item := &model.Item{
		Qualifier: ctx.NameToQualifier(name),
		Content: &model.ItemContent{
			Name:       name,
			Source:     src,
			PartNumber: pn,
		},
	}

	implAttr, ok := attrs["impl"]
	if ok {
		item.Implement, _ = c.resolveContractsLinesAttr(ctx, implAttr)
	}

	item.Digest, err = digest.SHA256FromSymbol(item)
	if err != nil {
		return item, err
	}

	err = c.buildCompanionForItem(ctx, item, input)
	if err != nil {
		return nil, err
	}

	item.Content.Details, err = c.getDetails(ctx, attrs)
	return item, err
}

func (c *Core) parseCoItemBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.CoItem, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	pn, _ := getString(attrs, "part_number")
	// ref, _ := getString(attrs, "ref")
	src, _ := getString(attrs, "source")

	var err error

	coItem := &model.CoItem{
		Qualifier: ctx.NameToQualifier(name),
		Content: &model.ItemContent{
			Name:       name,
			Source:     src,
			PartNumber: pn,
		},
	}

	reqAttr, ok := attrs["req"]
	if ok {
		coItem.Require, _ = c.resolveContractsLinesAttr(ctx, reqAttr)
	}

	coItem.Digest, err = digest.SHA256FromSymbol(coItem)
	return coItem, err
}

func (c *Core) createProcessContract(process *model.Process, mode string, line *UnresolvedBOMLine) (*model.Contract, error) {
	ret := &model.Contract{
		Qualifier: fmt.Sprintf("%s.%s.%s", process.Qualifier, mode, line.Role),
	}
	return ret, nil
}

func (c *Core) resolveBOMLinesAttr(ctx *ParserContext, attr *hcl.Attribute) ([]*model.BOMLine, error) {
	lines := parseBOMLinesAttr(ctx, attr)
	ret := make([]*model.BOMLine, 0, len(lines))
	for _, line := range lines {
		resolved, err := c.ResolveBOMLine(ctx, line)
		if err != nil {
			return nil, err
		}
		ret = append(ret, resolved)
	}
	return ret, nil
}

func (c *Core) parseProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (ret *model.Process, err error) {
	name := block.Labels[0]
	ret = &model.Process{
		Qualifier: ctx.NameToQualifier(name),
	}
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	cycle, _ := getNumber(attrs, "cycle")
	ret.CycleTime = cycle
	ret.Input, err = c.resolveBOMLinesAttr(ctx, attrs["input"])
	if err != nil {
		return
	}
	ret.Output, err = c.resolveBOMLinesAttr(ctx, attrs["output"])
	if err != nil {
		return
	}
	ret.Digest, err = digest.SHA256FromSymbol(ret)
	return
}

func (c *Core) parseCoProcessBlock(ctx *ParserContext, block *hclsyntax.Block) (ret *model.CoProcess, err error) {
	name := block.Labels[0]
	ret = &model.CoProcess{
		Qualifier: ctx.NameToQualifier(name),
	}
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	cycle, _ := getNumber(attrs, "cycle")
	ret.CycleTime = cycle
	ret.Input, err = c.resolveBOMLinesAttr(ctx, attrs["input"])
	if err != nil {
		return
	}
	ret.Output, err = c.resolveBOMLinesAttr(ctx, attrs["output"])
	if err != nil {
		return
	}
	ret.Digest, err = digest.SHA256FromSymbol(ret)
	return
}

func (c *Core) parseContractBlock(ctx *ParserContext, block *hclsyntax.Block) (*model.Contract, error) {
	name := block.Labels[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	params := make(map[string]any)
	for attr := range attrs {
		// TODO: support more types?
		param, _ := getString(attrs, attr)
		params[attr] = param
	}
	contract := &model.Contract{
		Name:      name,
		Qualifier: ctx.NameToQualifier(name),
		Params:    params,
	}
	digest, err := digest.SHA256FromSymbol(contract)
	if err != nil {
		return contract, err
	}
	contract.Digest = digest
	return contract, nil
}
