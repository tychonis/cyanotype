package bom

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/parser/hcl"
	"github.com/tychonis/cyanotype/model"
)

var Cmd = &cobra.Command{
	Use:   "bom [filename] [bom root]",
	Short: "generate bom from bpo",
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
	rootSymbol, err := core.Symbols.Resolve(rootPath)
	if err != nil {
		slog.Warn("Root item not found.", "item", rootPart)
		return
	}
	rootItem, ok := rootSymbol.(model.BOMItem)
	if !ok {
		slog.Warn("Root symbol not resolved.", "item", rootPart)
	}

	counter := core.Count(rootPath)
	fmt.Printf("Part usage in %s (Part #: %s):\n", rootPart, rootItem.GetPartNumber())
	// for name, qty := range counter {
	// 	fmt.Printf("- %s (Part %v #: %s): %1f\n",
	// 		name,
	// 		bom.Items[name].GetID(),
	// 		bom.Items[name].GetPartNumber(),
	// 		qty,
	// 	)
	// }
	core.CounterToCSV(counter)
}
