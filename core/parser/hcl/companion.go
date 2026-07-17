package hcl

import (
	"errors"
	"log/slog"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/core/qualifier"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) buildCompanionCoItem(item *model.Item) (*model.CoItem, error) {
	var err error
	co := &model.CoItem{}
	co.Type = "coitem"
	co.Qualifier = qualifier.ImplicitCoItem(item)
	co.Digest, err = digest.SHA256FromSymbol(co)
	if err != nil {
		return co, err
	}
	return co, p.Symbols.RegisterConcreteSymbol(co)
}

func (p *Parser) buildCompanionCoProcess(item *model.Item, coItem *model.CoItem) (*process.CoProcess, error) {
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

	cp := &process.CoProcess{}
	cp.Type = "coprocess"
	cp.Qualifier = qualifier.ImplicitCoProcess(item)
	cp.Content = content
	cp.Digest, err = digest.SHA256FromSymbol(cp)
	if err != nil {
		return cp, err
	}
	return cp, p.Symbols.RegisterConcreteSymbol(cp)
}

func (p *Parser) buildCompanionProcess(item *model.Item, pc process.ProcessContent) (*process.Process, error) {
	var err error
	switch content := pc.(type) {
	case *process.Abstract:
		content.Output = []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
			},
		}
	case *process.Drawing:
		content.Output = []*model.BOMLine{
			{
				Item: item.Digest,
				Qty:  1,
			},
		}
	default:
		return nil, errors.New("process content type not recognized")
	}

	process := &process.Process{}
	process.Type = "process"
	process.Qualifier = qualifier.ImplicitProcess(item)
	process.Content = pc
	process.Digest, err = digest.SHA256FromSymbol(process)
	if err != nil {
		return process, err
	}
	return process, p.Symbols.RegisterConcreteSymbol(process)
}

type Companion struct {
	CoItem    *model.CoItem
	Process   *process.Process
	CoProcess *process.CoProcess
}

func (p *Parser) buildCompanionForItem(ctx *ParserContext, item *model.Item, pc process.ProcessContent) (*Companion, error) {
	slog.Debug("build companions", "module", ctx.CurrentModule(), "item", item.Qualifier)

	companion := &Companion{}

	coItem, err := p.buildCompanionCoItem(item)
	if err != nil {
		return companion, err
	}
	companion.CoItem = coItem

	coProcess, err := p.buildCompanionCoProcess(item, coItem)
	if err != nil {
		return companion, err
	}
	companion.CoProcess = coProcess

	process, err := p.buildCompanionProcess(item, pc)
	if err != nil {
		return companion, err
	}
	companion.Process = process
	return companion, nil
}
