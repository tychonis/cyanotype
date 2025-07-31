package model

type Contract struct {
	Qualifier string `json:"qualifier" yaml:"qualifier"`
	Name      string `json:"name" yaml:"name"`
	Params    map[string]any
}
