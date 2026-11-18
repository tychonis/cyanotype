package model

import "github.com/tychonis/cyanotype/internal/stable"

type Qualifier = string

type Resolver interface {
	Resolve(path []string) (Resolver, error)
}

type Typed interface {
	GetType() string
}

type ConcreteSymbol interface {
	Typed
	GetQualifier() Qualifier
	GetDigest() Digest
	SetQualifier(Qualifier)
	SetDigest(Digest)
}

type SymbolBase struct {
	Qualifier Qualifier `json:"-" yaml:"-"`
	Digest    Digest    `json:"-" yaml:"-"`
}

func (s *SymbolBase) SetQualifier(q Qualifier) {
	s.Qualifier = q
}

func (s *SymbolBase) SetDigest(d Digest) {
	s.Digest = d
}

func (s *SymbolBase) GetQualifier() Qualifier {
	return s.Qualifier
}

func (s *SymbolBase) GetDigest() Digest {
	return s.Digest
}

type SymbolContent struct {
	Name      string      `json:"name" yaml:"name"`
	Artifacts []*Artifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
	Details   stable.Map  `json:"details" yaml:"details"`
}

func (s *SymbolContent) Set(key string, val any) {
	if s.Details == nil {
		s.Details = make(stable.Map)
	}
	s.Details[key] = val
}

func (s *SymbolContent) Get(key string) any {
	if s.Details == nil {
		return nil
	}
	return s.Details[key]
}

type BOMLine struct {
	Name string  `json:"name" yaml:"name"`
	Item ItemID  `json:"item" yaml:"item"`
	Qty  float64 `json:"qty" yaml:"qty"`
}
