package model

import (
	"errors"

	"github.com/tychonis/cyanotype/internal/stable"
)

type ItemID = Digest
type LinkType string

type ItemBase struct {
	Type      string       `json:"type" yaml:"type"`
	Qualifier string       `json:"qualifier" yaml:"qualifier"`
	Content   *ItemContent `json:"content" yaml:"content"`
	Digest    ItemID       `json:"-" yaml:"-"`
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

// TODO: implement attrs?
func (i *Item) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return i, nil
}

func (i *Item) GetQualifier() string {
	return i.Qualifier
}

func (i *Item) GetDigest() string {
	return i.Digest
}

func (i *Item) GetType() string {
	return i.Type
}

// TODO: implement attrs?
func (ci *CoItem) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return ci, nil
}

func (ci *CoItem) GetQualifier() string {
	return ci.Qualifier
}

func (ci *CoItem) GetDigest() string {
	return ci.Digest
}

func (ci *CoItem) GetType() string {
	return ci.Type
}
