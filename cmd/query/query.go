package query

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/serializer"
)

var Cmd = &cobra.Command{
	Use:   "query <symbol>",
	Short: "Query queries the local catalog for a given symbol",
	RunE:  run,
	Args:  cobra.ExactArgs(1),
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

func run(cmd *cobra.Command, args []string) error {
	qualifier := args[0]

	cat := catalog.New("local")
	sym, err := cat.FindCurrent(qualifier)
	if err != nil {
		slog.Error("Failed to find item.", "error", err)
		return err
	}
	fmt.Print(sym.GetDigest() + ":")
	report(sym)

	meta, err := cat.GetMetadata(sym.GetDigest())
	if err != nil {
		slog.Error("Failed to get metadata.", "error", err)
		return err
	}
	report(meta)
	return nil
}
