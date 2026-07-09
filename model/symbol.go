package model

type Qualifier = string

type Symbol interface {
	Resolve(path []string) (Symbol, error)
}

type ConcreteSymbol interface {
	Symbol
	GetType() string
	GetQualifier() Qualifier
	GetDigest() Digest
}

type BOMLine struct {
	Name string  `json:"name" yaml:"name"`
	Item ItemID  `json:"item" yaml:"item"`
	Qty  float64 `json:"qty" yaml:"qty"`
}
