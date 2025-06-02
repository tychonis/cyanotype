package build

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "build [filename] [target]",
	Short: "build bpc from bpo",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	bpcPath := args[1]

	bom := hcl.Parse(bpoPath)

	output, _ := os.Create(bpcPath)
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	encoder.Encode(bom.Catalog)
}
