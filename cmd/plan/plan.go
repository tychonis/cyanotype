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
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	if bpoPath == "" {
		bpoPath = "."
	}

	core := hcl.NewParser()
	err := core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	cat := catalog.New("local")
	err = core.PreviewCommit(cat)
	if err != nil {
		slog.Error("Failed to commit to catalog.", "error", err)
		return
	}
}
