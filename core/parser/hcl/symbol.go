package hcl

import (
	"errors"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/tychonis/cyanotype/model"
)

type SymbolAttribute struct {
	SymbolQualifier string
	RemainingPath   []string
}

func (sa *SymbolAttribute) Resolve(path []string) (model.Resolver, error) {
	if len(path) == 0 {
		return sa, nil
	}
	return nil, errors.New("cannot resolve symbol attribute with remaining path")
}

type UnprocessedSymbol struct {
	Context *ParserContext
	Block   *hclsyntax.Block

	qualifier string
}

func (us *UnprocessedSymbol) Resolve(path []string) (model.Resolver, error) {
	if len(path) == 0 {
		return us, nil
	}
	return &SymbolAttribute{
		SymbolQualifier: us.qualifier,
		RemainingPath:   path,
	}, nil
}
