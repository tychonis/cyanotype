package model

import (
	"github.com/google/uuid"
)

type ItemID = uuid.UUID
type NodeID = uuid.UUID
type LinkType string

const (
	DERIVE_SUPERSESSEION   LinkType = "supersession"
	DERIVE_INTERCHANGEABLE LinkType = "interchangeable"
	DERIVE_VARIANT         LinkType = "variant"
	DERIVE_CHANGE          LinkType = "change"
)

// Item corresponds to a single immutable snapshot of a part or assembly.
// Any change to spec, composition, process or metadata produces a new Item.
// Items are linked through Predecessor for traceability supersession or interchangeability.
type Item struct {
	ID        ItemID `json:"id" yaml:"id"`
	Qualifier string `json:"qualifier" yaml:"qualifier"`

	Derivation *Derivation `json:"derivation" yaml:"derivation"`
	Process    ProcessID   `json:"process" yaml:"process"`

	Require   []ContractID `json:"require" yaml:"require"`
	Implement []ContractID `json:"implement" yaml:"implement"`

	Content *ItemContent `json:"content" yaml:"content"`

	Digest string `json:"-" yaml:"-"`
}

type ItemContent struct {
	Name       string       `json:"name" yaml:"name"`
	Source     string       `json:"source,omitempty" yaml:"source,omitempty"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	References []*Reference `json:"ref,omitempty" yaml:"ref,omitempty"`
}

type Reference struct {
	Reference string `json:"ref" yaml:"ref"`
	Tag       string `json:"tag" yaml:"tag"`
	Path      string `json:"path" yaml:"path"`
	Digest    string `json:"digest" yaml:"digest"`
}

type ItemNode struct {
	ID       NodeID   `json:"id" yaml:"id"`
	Path     string   `json:"path" yaml:"path"`
	ItemID   ItemID   `json:"item_id" yaml:"item_id"`
	ParentID NodeID   `json:"parent_id" yaml:"parent_id"`
	Children []NodeID `json:"children" yaml:"children"`
	Qty      float64  `json:"qty" yaml:"qty"`
}

type Derivation struct {
	DerivedFrom ItemID
	Type        LinkType
}
