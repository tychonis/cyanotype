package hcl

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/model"
)

type UnprocessedSymbol struct {
	Context *ParserContext
	Block   *hclsyntax.Block

	qualifier string
}

func (us *UnprocessedSymbol) Resolve(path []string) (model.Symbol, error) {
	return us, nil
}

type UnresolvedBOMLine struct {
	Role string   `json:"role" yaml:"role"`
	Ref  []string `json:"ref" yaml:"ref"`
	Qty  float64  `json:"qty" yaml:"qty"`
}

type ItemSyntaxSugar struct {
	From []*UnresolvedBOMLine
}

// itemSymbol is the result of first pass, all symbols are not linked.
type ItemSymbol struct {
	Qualifier string
	Implement []string

	Content     *model.ItemContent
	SyntaxSugar *ItemSyntaxSugar
}

func (is *ItemSymbol) Resolve(path []string) (model.Symbol, error) {
	return nil, nil
}

type ProcessSymbol struct {
	Qualifier string
	CycleTime float64
	Input     []*UnresolvedBOMLine
	Output    []*UnresolvedBOMLine
}

func (ps *ProcessSymbol) Resolve(path []string) (model.Symbol, error) {
	return nil, nil
}
