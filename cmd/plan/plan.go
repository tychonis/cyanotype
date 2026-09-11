package plan

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan shows the diff between working-tree and local catalog",
	RunE:  run,
}
var ignoreArtifacts bool

func init() {
	Cmd.Flags().BoolVar(&ignoreArtifacts, "ignore-artifacts", false, "ignore artifacts during commit")
}

func run(cmd *cobra.Command, args []string) error {
	bpoPath := "."

	p := hcl.NewParser()
	p.Options.IgnoreArtifacts = ignoreArtifacts
	err := p.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return err
	}

	cat := catalog.New("local")
	err = p.PreviewCommit(cat)
	if err != nil {
		slog.Error("Failed to commit to catalog.", "error", err)
		return err
	}
	return nil
}
