package push

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "push [filename]",
	Short: "Adhoc implementation saving catalog to remote",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	if bpoPath == "" {
		bpoPath = "."
	}

	server := args[1]
	tag := args[2]

	core := hcl.NewCore("memory")
	err := core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	err = core.SaveCatalog(server, tag)
	if err != nil {
		slog.Warn("Failed to save catalog.", "error", err)
		return
	}
}
