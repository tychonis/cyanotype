package bom

import (
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "bom [filename] [bom root]",
	Short: "Generate bom from bpo",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func init() {
	Cmd.Flags().StringP("output", "o", "csv", "set output format")
}

func run(cmd *cobra.Command, args []string) {
	bomPath := args[0]
	rootPart := args[1]

	outputFmt := cmd.Flag("output").Value.String()
	if outputFmt != "csv" {
		slog.Warn("Format not supported.", "format", outputFmt)
	}

	core := hcl.NewCore()
	err := core.Parse(bomPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	rootPath := strings.Split(rootPart, ".")

	counter := core.Count(rootPath)
	core.CounterToCSV(counter)
}
