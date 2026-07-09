package catalog

import "github.com/tychonis/cyanotype/model"

type Metadata struct {
	IntroducedBy model.RevisionID `json:"introduced_by"`
}
