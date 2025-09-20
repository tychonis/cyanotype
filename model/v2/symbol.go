package model

import "github.com/tychonis/cyanotype/model"

type Symbol interface {
	Resolve(path []string) (model.Symbol, error)
}
