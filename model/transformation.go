package model

type TransformationID = Digest

type Transformation struct {
	Qualifier string `json:"qualifier" yaml:"qualifier"`

	Digest TransformationID `json:"-" yaml:"-"`
}
