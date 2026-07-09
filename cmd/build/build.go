package build

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "build <path>",
	Short: "Build revision from bpo, report errors but don't commit to catalog",
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
}
