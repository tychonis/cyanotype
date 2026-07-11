package query

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/serializer"
)

var Cmd = &cobra.Command{
	Use:   "query",
	Short: "Query queries the local catalog for a given symbol",
	Run:   run,
}

func report(data any) error {
	bytes, err := serializer.Serialize(data)
	if err != nil {
		slog.Error("Failed to serialize data.", "error", err)
		return err
	}
	fmt.Print(string(bytes) + "\n")
	return nil
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	qualifier := args[1]
	if bpoPath == "" {
		bpoPath = "."
	}

	cat := catalog.New("local")
	sym, err := cat.FindCurrent(qualifier)
	if err != nil {
		slog.Error("Failed to find item.", "error", err)
		return
	}
	fmt.Print(sym.GetDigest() + ":")
	report(sym)

	meta, err := cat.GetMetadata(sym.GetDigest())
	if err != nil {
		slog.Error("Failed to get metadata.", "error", err)
		return
	}
	report(meta)
}
