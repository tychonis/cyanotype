package build

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/internal/parser/hcl"
	"github.com/tychonis/cyanotype/model"
)

var Cmd = &cobra.Command{
	Use:   "build [filename] [bom root]",
	Short: "Build bpc from bpo",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func init() {
	// TODO: distinguish from output format
	Cmd.Flags().StringP("output", "o", "", "set output path")
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	root := args[1]

	bpcPath := cmd.Flag("output").Value.String()
	if bpcPath == "" {
		bpcPath = strings.ReplaceAll(bpoPath, ".bpo", ".bpc")
		// Folder
		if !strings.Contains(bpcPath, ".bpc") {
			bpcPath = "ouptput.bpc"
		}
	}
	core := hcl.NewCore("local")
	err := core.Process(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	rootSym, err := core.Catalog.Find(root)
	if err != nil {
		slog.Error("Failed to find root item.", "error", err)
		return
	}

	rootItem, ok := rootSym.(*model.Item)
	if !ok {
		slog.Error("Root is not an Item.")
		return
	}

	rootNode, err := core.Build(rootItem)
	if err != nil {
		slog.Error("Failed to build.", "error", err)
		return
	}

	output := rootNode.Export()
	os.WriteFile(bpcPath, output, 0o644)
}
