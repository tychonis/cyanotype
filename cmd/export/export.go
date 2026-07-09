package export

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "export <path> <root>",
	Short: "Export catalog from bpo",
	Run:   run,
}

func init() {
	// TODO: distinguish from output format
	Cmd.Flags().StringP("output", "o", "", "set output path")
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]

	catalogPath := cmd.Flag("output").Value.String()
	if catalogPath == "" {
		catalogPath = strings.ReplaceAll(bpoPath, ".bpo", ".bpc")
		// Folder
		if !strings.Contains(catalogPath, ".bpc") {
			catalogPath = "catalog.json"
		}
	}
	cat := catalog.New("local")
	slog.Debug("catalog", "catalog", cat)

	output, err := cat.Export()
	if err != nil {
		slog.Error("Failed to export catalog.", "error", err)
		return
	}
	os.WriteFile(catalogPath, output, 0o644)
}
