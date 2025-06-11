package symbols

import (
	"errors"

	"github.com/tychonis/cyanotype/model"
)

type SymbolTable struct {
	Modules map[string]*ModuleScope
}

type ModuleScope struct {
	Symbols map[string]model.Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{Modules: make(map[string]*ModuleScope)}
}

func NewModuleScope() *ModuleScope {
	return &ModuleScope{Symbols: make(map[string]model.Symbol)}
}

func (t *SymbolTable) AddSymbol(module string, name string, symbol model.Symbol) error {
	if t.Modules[module] == nil {
		t.Modules[module] = NewModuleScope()
	}
	_, ok := t.Modules[module].Symbols[name]
	if ok {
		return errors.New("symbol existed")
	}
	t.Modules[module].Symbols[name] = symbol
	return nil
}
