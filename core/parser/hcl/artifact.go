package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

func ParseArtifacts(ctx *ParserContext, block *hclsyntax.Block) ([]*model.Artifact, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}

	var artifacts []*model.Artifact

	for _, child := range block.Body.Blocks {
		if child.Type != "artifact" {
			continue
		}

		artifact, err := parseArtifactBlock(ctx, child)
		if err != nil {
			return nil, err
		}

		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

func parseArtifactBlock(_ *ParserContext, block *hclsyntax.Block) (*model.Artifact, error) {
	if len(block.Labels) != 1 {
		return nil, fmt.Errorf(
			"%s: artifact block must have exactly one label",
			block.TypeRange.String(),
		)
	}

	artifact := &model.Artifact{
		Name: block.Labels[0],
	}

	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}

	path, err := getString(attrs, "path")
	if err != nil {
		return nil, err
	}
	artifact.Path = path
	artifact.Digest, err = digest.SHA256FromFile(path)
	if err != nil {
		return nil, err
	}
	tag, err := getString(attrs, "tag")
	if err != nil {
		return nil, err
	}
	artifact.Tag = tag
	return artifact, nil
}
