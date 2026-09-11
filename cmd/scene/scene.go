package scene

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/scene"
	"github.com/tychonis/cyanotype/model"
)

var Cmd = &cobra.Command{
	Use:   "scene <item>",
	Short: "Generate scene from item",
	RunE:  run,
	Args:  cobra.ExactArgs(1),
}

func run(cmd *cobra.Command, args []string) error {
	qualifier := args[0]

	cat := catalog.New("local")
	sym, err := cat.FindCurrent(qualifier)
	if err != nil {
		slog.Error("Failed to find item.", "error", err)
		return err
	}
	item, ok := sym.(*model.Item)
	if !ok {
		slog.Error("Failed to assert symbol as item.")
		return fmt.Errorf("failed to assert symbol as item")
	}

	scene.RenderSymbol(item)

	// meta, err := cat.GetMetadata(sym.GetDigest())
	// if err != nil {
	// 	slog.Error("Failed to get metadata.", "error", err)
	// 	os.Exit(1)
	// }
	return nil
}
