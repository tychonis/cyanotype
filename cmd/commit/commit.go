package commit

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/internal/parser/hcl/v2"
)

var Cmd = &cobra.Command{
	Use:   "commit",
	Short: "Build bpc from bpo",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func init() {
	// TODO: distinguish from output format
	Cmd.Flags().StringP("output", "o", "", "set output path")
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := "."

	core := hcl.NewCore()
	err := core.Parse(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	err = core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to build bom graph.", "error", err)
	}
}
