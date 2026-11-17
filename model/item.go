package model

type ItemID = Digest

type ItemBase struct {
	SymbolBase

	Content *SymbolContent `json:"content" yaml:"content"`
}

// Item corresponds to a single immutable snapshot of a part or assembly.
// Any change to spec, composition, process or metadata produces a new Item.
// Items are linked through Predecessor for traceability supersession or interchangeability.
type Item struct {
	ItemBase
	Implement []ContractID `json:"implement" yaml:"implement"`
}

// CoItem defines requirements.
type CoItem struct {
	ItemBase
	Require []ContractID `json:"require" yaml:"require"`
}

func (i *Item) GetType() string {
	return "item"
}

func (i *Item) GetName() string {
	if i.Content != nil {
		return i.Content.Name
	}
	return i.GetQualifier()
}

func (ci *CoItem) GetType() string {
	return "coitem"
}

func (ci *CoItem) GetName() string {
	if ci.Content != nil {
		return ci.Content.Name
	}
	return ci.GetQualifier()
}
