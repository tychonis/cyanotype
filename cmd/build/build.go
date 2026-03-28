package build

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/parser/hcl"
	"github.com/tychonis/cyanotype/internal/catalog"
)

var Cmd = &cobra.Command{
	Use:   "build [filename]",
	Short: "Build .bpc folder from bpo",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	if bpoPath == "" {
		bpoPath = "."
	}

	catalog.Initialize()

	core := hcl.NewCore("local")
	err := core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}
}
