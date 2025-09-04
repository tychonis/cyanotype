package model

type Symbol interface {
	Resolve(path []string) (Symbol, error)
}
