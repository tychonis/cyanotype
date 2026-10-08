package catalog

import "github.com/tychonis/cyanotype/model"

func (c *Catalog) ListArtifacts() ([]*model.Artifact, error) {
	syms, err := c.GetSymbols()
	if err != nil {
		return nil, err
	}
	ret := make([]*model.Artifact, 0)
	for _, sym := range syms {
		switch sym := sym.(type) {
		case *model.Item:
			ret = append(ret, sym.Content.Artifacts...)
		case *model.CoItem:
			ret = append(ret, sym.Content.Artifacts...)
		default:
			continue
		}
	}
	return ret, nil
}
