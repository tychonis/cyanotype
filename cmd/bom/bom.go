package bom

import (
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/instantiator"
	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "bom <path> <root>",
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

	p := hcl.NewParser()
	err := p.Build(bomPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	cat := catalog.New("memory")
	err = p.Commit(cat)
	if err != nil {
		slog.Error("Failed to commit to catalog.", "error", err)
		return
	}
	ins := instantiator.New()

	counter, err := ins.Count(cat, rootPart)
	if err != nil {
		slog.Warn("Error counting", "error", err)
	}
	ins.CounterToCSV(counter)
}
