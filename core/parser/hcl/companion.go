package hcl

import (
	"log/slog"

	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func (c *Core) buildCompanionCoItem(item *model.Item) (*model.CoItem, error) {
	var err error
	co := &model.CoItem{
		Qualifier: getImplicitCoItemQualifier(item),
	}
	co.Digest, err = digest.SHA256FromSymbol(co)
	if err != nil {
		return co, err
	}
	return co, c.Catalog.Add(co)
}

func (c *Core) buildCompanionCoProcess(item *model.Item, coItem *model.CoItem) (*model.CoProcess, error) {
	var err error
	cp := &model.CoProcess{
		Qualifier: getImplicitCoProcessQualifier(item),
		Input: []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
			},
		},
		Output: []*model.BOMLine{
			{
				Item: coItem.Digest,
				Qty:  1,
			},
		},
	}
	cp.Digest, err = digest.SHA256FromSymbol(cp)
	if err != nil {
		return cp, err
	}
	return cp, c.Catalog.Add(cp)
}

func (c *Core) buildCompanionProcess(item *model.Item, input []*model.BOMLine) (*model.Process, error) {
	var err error
	p := &model.Process{
		Qualifier: getImplicitProcessQualifier(item),
		Output: []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
				Role: DEFAULT,
			},
		},
		Input: input,
	}
	p.Digest, err = digest.SHA256FromSymbol(p)
	if err == nil {
		c.Catalog.Add(p)
	}
	return p, err
}

func (c *Core) buildCompanionForItem(ctx *ParserContext, item *model.Item, input []*model.BOMLine) error {
	slog.Debug("build companions", "module", ctx.CurrentModule(), "item", item.Qualifier)

	co, err := c.buildCompanionCoItem(item)
	if err != nil {
		return err
	}

	_, err = c.buildCompanionCoProcess(item, co)
	if err != nil {
		return err
	}

	_, err = c.buildCompanionProcess(item, input)
	return err
}
