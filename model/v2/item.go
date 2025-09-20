package model

import (
	"errors"

	"github.com/tychonis/cyanotype/internal/stable"
	"github.com/tychonis/cyanotype/model"
)

type ItemID = Digest
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
	Qualifier string `json:"qualifier" yaml:"qualifier"`

	Derivation *Derivation `json:"derivation" yaml:"derivation"`
	Process    ProcessID   `json:"process" yaml:"process"`

	Require   []ContractID `json:"require" yaml:"require"`
	Implement []ContractID `json:"implement" yaml:"implement"`

	Content *ItemContent `json:"content" yaml:"content"`

	Digest ItemID `json:"-" yaml:"-"`
}

type ItemContent struct {
	Name       string       `json:"name" yaml:"name"`
	Source     string       `json:"source,omitempty" yaml:"source,omitempty"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	References []*Reference `json:"ref,omitempty" yaml:"ref,omitempty"`
	Details    stable.Map   `json:"details" yaml:"details"`
}

type Reference struct {
	Reference string `json:"ref" yaml:"ref"`
	Tag       string `json:"tag" yaml:"tag"`
	Path      string `json:"path" yaml:"path"`
	Digest    string `json:"digest" yaml:"digest"`
}

type Derivation struct {
	DerivedFrom ItemID
	Type        LinkType
}

// TODO: implement attrs?
func (i *Item) Resolve(path []string) (model.Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return i, nil
}
