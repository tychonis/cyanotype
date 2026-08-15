package hcl

import (
	"errors"
	"log/slog"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/model"
)

func (p *Parser) Commit(cat *catalog.Catalog) error {
	return p.commit(cat, false)
}

func (p *Parser) PreviewCommit(cat *catalog.Catalog) error {
	return p.commit(cat, true)
}

func (p *Parser) isSymbolChanged(old model.ConcreteSymbol, new model.ConcreteSymbol) bool {
	if old == nil && new == nil {
		return false
	}
	if old == nil || new == nil {
		return true
	}
	return old.GetDigest() != new.GetDigest()
}

func (p *Parser) commit(cat *catalog.Catalog, dryrun bool) error {
	revision := cat.NewRevision()
	change := 0
	for qualifier, symDigest := range p.Symbols.QualifierIndex {
		oldSym, err := cat.FindCurrent(qualifier)
		if err != nil && err != catalog.ErrNotFound {
			return err
		}
		newSym, ok := p.Symbols.ConcreteSymbols[symDigest]
		if !ok {
			return errors.New("symbol not found in symbol table")
		}
		if !p.isSymbolChanged(oldSym, newSym) {
			continue
		}
		if !dryrun {
			err = cat.Add(revision, newSym)
			if err != nil {
				return err
			}
		} else {
			slog.Info("New symbol", "qualifier", qualifier, "digest", symDigest)
		}
		change++
	}
	if dryrun || change == 0 {
		return nil
	}
	return cat.Commit(revision)
}
