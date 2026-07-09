package hcl

import (
	"errors"

	"github.com/tychonis/cyanotype/core/catalog"
)

func (p *Parser) Commit(cat *catalog.Catalog) error {
	revision := cat.NewRevision()
	change := 0
	for qualifier, symDigest := range p.Symbols.QualifierIndex {
		oldSym, err := cat.FindCurrent(qualifier)
		if err != nil && err != catalog.ErrNotFound {
			return err
		}
		if oldSym != nil && oldSym.GetDigest() == symDigest {
			continue
		}
		sym, ok := p.Symbols.ConcreteSymbols[symDigest]
		if !ok {
			return errors.New("symbol not found in symbol table")
		}
		cat.Add(revision, sym)
		change++
	}
	if change == 0 {
		return nil
	}
	cat.Commit(revision)
	return nil
}
