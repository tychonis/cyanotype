package symbols

import (
	"fmt"

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
		return fmt.Errorf("symbol %s already existed in %s", name, module)
	}
	t.Modules[module].Symbols[name] = symbol
	return nil
}

func (t *SymbolTable) Resolve(ref []string) (model.Symbol, error) {
	mod, ok := t.Modules["."]
	if !ok {
		return nil, fmt.Errorf("no registered symbols")
	}
	return mod.Resolve(ref)
}

func (m *ModuleScope) Resolve(ref []string) (model.Symbol, error) {
	if len(ref) <= 0 {
		return nil, fmt.Errorf("resolving empty symbol")
	}
	sym := ref[0]
	resolver, ok := m.Symbols[sym]
	if !ok {
		return nil, fmt.Errorf("symbol %v not registered", ref)
	}
	return resolver.Resolve(ref[1:])
}
