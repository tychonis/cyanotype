package hcl

import (
	"errors"
	"log/slog"

	"github.com/tychonis/cyanotype/internal/catalog"
	"github.com/tychonis/cyanotype/model"
)

type UnresolvedBOMLine struct {
	Role string   `json:"role" yaml:"role"`
	Ref  []string `json:"ref" yaml:"ref"`
	Qty  float64  `json:"qty" yaml:"qty"`
}

func (c *Core) processKeywordFROM(ctx *ParserContext, from []*UnresolvedBOMLine) ([]*model.BOMLine, error) {
	if len(from) <= 0 {
		return nil, nil
	}
	ret := make([]*model.BOMLine, 0, len(from))
	for _, comp := range from {
		qualifier := refToQualifier(ctx, comp.Ref)
		compItemSym, err := c.Catalog.Find(qualifier)
		if err != nil {
			if err != catalog.ErrNotFound {
				return nil, err
			} else {
				sym, err := c.Resolve(ctx, comp.Ref)
				if err != nil {
					return nil, err
				}
				unprocessed, ok := sym.(*UnprocessedSymbol)
				if !ok {
					return nil, errors.New("wrong symbol type")
				}
				compItemSym, err = c.ParseSymbol(unprocessed)
				if err != nil {
					return nil, err
				}
			}
		}

		compItem, ok := compItemSym.(*model.Item)
		if !ok {
			return nil, errors.New("incorrect ref")
		}
		compCoItems, err := c.Catalog.GetCoItems(compItem.Digest)
		if err != nil {
			return nil, err
		}
		if len(compCoItems) != 1 {
			slog.Debug("error", "item", compItem.Qualifier, "length", len(compCoItems), "digest", compItem.Digest)
			return nil, errors.New("not implemented yet")
		}
		ret = append(ret, &model.BOMLine{
			Item: compCoItems[0].Item,
			Qty:  comp.Qty,
		})
	}
	return ret, nil
}
