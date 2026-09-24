package symbol

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list <file>",
	Short: "List symbols in a file",
	RunE:  runList,
	Args:  cobra.ExactArgs(1),
}

func runList(cmd *cobra.Command, args []string) error {
	file := args[0]

	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCLFile(file)
	if diags.HasErrors() {
		slog.Error("Failed to parse file.", "error", diags.Error())
		return diags
	}

	content, ok := hclFile.Body.(*hclsyntax.Body)
	if !ok {
		slog.Error("Failed to parse content.")
		return errors.New("failed to parse content")
	}

	for _, block := range content.Blocks {
		ext := ""
		for name, attr := range block.Body.Attributes {
			if name == "ext" {
				extVal, _ := attr.Expr.Value(nil)
				ext = extVal.AsString()
			}
		}
		fmt.Printf("%s %s %s\n", block.Type, block.Labels[0], ext)
	}
	return nil
}
