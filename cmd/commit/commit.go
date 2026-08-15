package commit

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "commit",
	Short: "Build bpc from bpo",
	Run:   run,
}
var ignoreArtifacts bool

func init() {
	// TODO: distinguish from output format
	Cmd.Flags().StringP("output", "o", "", "set output path")
	Cmd.Flags().BoolVar(&ignoreArtifacts, "ignore-artifacts", false, "ignore artifacts during commit")
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := "."

	p := hcl.NewParser()
	p.Options.IgnoreArtifacts = ignoreArtifacts
	err := p.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	cat := catalog.New("local")
	err = p.Commit(cat)
	if err != nil {
		slog.Error("Failed to commit to catalog.", "error", err)
		return
	}
}
