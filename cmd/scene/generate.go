package scene

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/core/scene"
	"github.com/tychonis/cyanotype/model"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate <item>",
	Short: "Generate scene from item",
	RunE:  runGenerate,
	Args:  cobra.ExactArgs(1),
}

func init() {
	GenerateCmd.Flags().StringP("output", "o", "", "set output path")
}

func runGenerate(cmd *cobra.Command, args []string) error {
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

	rendered, err := scene.RenderSymbol(cat, item, true)
	if err != nil {
		slog.Error("Failed to render scene.", "error", err)
		return err
	}
	data, err := json.Marshal(rendered)
	if err != nil {
		slog.Error("Failed to marshal scene.", "error", err)
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	if output == "" {
		fmt.Println(string(data))
		return nil
	}

	old, err := os.ReadFile(output)
	if err == nil && bytes.Equal(old, data) {
		return nil
	}

	return os.WriteFile(output, data, 0644)
}
