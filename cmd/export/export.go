package export

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "export",
	Short: "Export catalog from .bpo",
	RunE:  run,
}

func init() {
	Cmd.Flags().StringP("output", "o", "", "set output path")
}

func run(cmd *cobra.Command, args []string) error {
	catalogPath := cmd.Flag("output").Value.String()
	if catalogPath == "" {
		catalogPath = "catalog.json"
	}
	cat := catalog.New("local")

	output, err := cat.Export()
	if err != nil {
		slog.Error("Failed to export catalog.", "error", err)
		return err
	}
	os.WriteFile(catalogPath, output, 0o644)
	return nil
}
