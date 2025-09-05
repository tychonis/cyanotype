package build

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/parser/hcl"
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
	rootPart := args[1]

	bpcPath := cmd.Flag("output").Value.String()
	if bpcPath == "" {
		bpcPath = strings.ReplaceAll(bpoPath, ".bpo", ".bpc")
		// Folder
		if !strings.Contains(bpcPath, ".bpc") {
			bpcPath = "ouptput.bpc"
		}
	}
	core := hcl.NewCore()
	err := core.Parse(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return
	}

	rootPath := strings.Split(rootPart, ".")
	bomGraph, err := core.Build(bpoPath, rootPath)
	if err != nil {
		slog.Warn("Failed to build bom graph.", "error", err)
	}

	for _, state := range core.States {
		bomGraph = bomGraph.Reference(state)
	}

	output, _ := os.Create(bpcPath)
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	encoder.Encode(bomGraph)
}
