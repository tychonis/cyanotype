package hcl

import (
	"log/slog"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func (c *Core) buildCompanionCoItem(item *model.Item) (*model.CoItem, error) {
	var err error
	co := &model.CoItem{}
	co.Qualifier = getImplicitCoItemQualifier(item)
	co.Digest, err = digest.SHA256FromSymbol(co)
	if err != nil {
		return co, err
	}
	return co, c.Catalog.Add(co)
}

func (c *Core) buildCompanionCoProcess(item *model.Item, coItem *model.CoItem) (*model.CoProcess, error) {
	var err error
	content := &process.Abstract{
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

	cp := &model.CoProcess{}
	cp.Qualifier = getImplicitCoProcessQualifier(item)
	cp.Content = content
	cp.Digest, err = digest.SHA256FromSymbol(cp)
	if err != nil {
		return cp, err
	}
	return cp, c.Catalog.Add(cp)
}

func (c *Core) buildCompanionProcess(item *model.Item, input []*model.BOMLine) (*model.Process, error) {
	var err error
	content := &process.Abstract{
		Output: []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
				Role: DEFAULT,
			},
		},
		Input: input,
	}

	p := &model.Process{}
	p.Qualifier = getImplicitProcessQualifier(item)
	p.Content = content
	p.Digest, err = digest.SHA256FromSymbol(p)
	if err != nil {
		return p, err
	}
	return p, c.Catalog.Add(p)
}

func (c *Core) buildCompanionForItem(ctx *ParserContext, item *model.Item, input []*model.BOMLine) error {
	slog.Debug("build companions", "module", ctx.CurrentModule(), "item", item.Qualifier)

	coItem, err := c.buildCompanionCoItem(item)
	if err != nil {
		return err
	}

	_, err = c.buildCompanionCoProcess(item, coItem)
	if err != nil {
		return err
	}

	_, err = c.buildCompanionProcess(item, input)
	return err
}
