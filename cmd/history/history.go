package history

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "history <symbol>",
	Short: "History shows the history of a given symbol",
	RunE:  run,
	Args:  cobra.ExactArgs(1),
}

func report(data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to serialize data.", "error", err)
		return err
	}
	fmt.Print(string(bytes) + "\n")
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	qualifier := args[0]

	cat := catalog.New("local")
	syms, err := cat.FindAll(qualifier)
	if err != nil {
		slog.Error("Failed to find item.", "error", err)
		return err
	}
	for _, sym := range syms {
		fmt.Print(sym.GetDigest() + ":")
		report(sym)

		meta, err := cat.GetMetadata(sym.GetDigest())
		if err != nil {
			slog.Error("Failed to get metadata.", "error", err)
			return err
		}
		report(meta)
	}
	return nil
}
