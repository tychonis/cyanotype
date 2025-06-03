package bom

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "bom [filename] [bom root]",
	Short: "generate bom from bpo",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func run(cmd *cobra.Command, args []string) {
	bomPath := args[0]
	rootPart := args[1]

	bom, err := hcl.Parse(bomPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
	}

	rootItem, ok := bom.Items[rootPart]
	if !ok {
		slog.Warn("Root item not found.", "item", rootPart)
		return
	}

	counter := bom.Count(rootPart)
	fmt.Printf("Part usage in %s (Part #: %s):\n", rootPart, rootItem.GetPartNumber())
	// for name, qty := range counter {
	// 	fmt.Printf("- %s (Part %v #: %s): %1f\n",
	// 		name,
	// 		bom.Items[name].GetID(),
	// 		bom.Items[name].GetPartNumber(),
	// 		qty,
	// 	)
	// }
	bom.CounterToCSV(counter)
}
