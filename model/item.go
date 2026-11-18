package model

type ItemID = Digest

// Item corresponds to a single immutable snapshot of a part or assembly.
// Any change to spec, composition, process or metadata produces a new Item.
// Items are linked through Predecessor for traceability supersession or interchangeability.
type Item struct {
	SymbolBase
	SymbolContent
	Implement []ContractID `json:"implement" yaml:"implement"`
}

// CoItem defines requirements.
type CoItem struct {
	SymbolBase
	SymbolContent
	Require []ContractID `json:"require" yaml:"require"`
}

func (i *Item) GetType() string {
	return "item"
}

func (i *Item) GetName() string {
	return i.Name
}

func (ci *CoItem) GetType() string {
	return "coitem"
}

func (ci *CoItem) GetName() string {
	return ci.Name
}
