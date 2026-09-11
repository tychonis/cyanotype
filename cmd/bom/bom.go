package bom

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/instantiator"
	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "bom <root>",
	Short: "Generate bom from bpo",
	RunE:  run,
	Args:  cobra.ExactArgs(1),
}

func init() {
	Cmd.Flags().StringP("format", "f", "csv", "set output format: csv, list")
}

func run(cmd *cobra.Command, args []string) error {
	rootPart := args[0]

	outputFmt, _ := cmd.Flags().GetString("format")

	p := hcl.NewParser()
	err := p.Build(".")
	if err != nil {
		slog.Error("Failed to parse bpo.", "error", err)
		return err
	}

	cat := catalog.New("memory")
	err = p.Commit(cat)
	if err != nil {
		slog.Error("Failed to commit to catalog.", "error", err)
		return err
	}

	ins := instantiator.New()
	counter, err := ins.Count(cat, rootPart)
	if err != nil {
		slog.Error("Error counting", "error", err)
		return err
	}

	switch outputFmt {
	case "csv":
		FmtCounterAsCSV(counter)
	case "list":
		FmtCounterAsList(counter)
	default:
		slog.Error("Format not supported.", "format", outputFmt)
		return fmt.Errorf("unsupported format: %s", outputFmt)
	}
	return nil
}
