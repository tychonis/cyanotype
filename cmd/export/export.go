package export

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "export [filename] [bom root]",
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
	core := hcl.NewCore("memory")
	err := core.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	output, _ := core.ExportCatalog()
	os.WriteFile(catalogPath, output, 0o644)
}
