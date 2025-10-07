package model

type Symbol interface {
	Resolve(path []string) (Symbol, error)
}

type ConcreteSymbol interface {
	Symbol
	GetQualifier() string
	GetDigest() string
}
