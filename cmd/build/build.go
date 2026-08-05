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
	var bpoPath string
	if len(args) == 0 {
		bpoPath = "."
	} else {
		bpoPath = args[0]
	}

	core := hcl.NewParser()
	err := core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}
}
