package build

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/parser/hcl"
)

var Cmd = &cobra.Command{
	Use:   "build",
	Short: "Build revision from bpo, report errors but don't commit to catalog",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	bpoPath := "."

	p := hcl.NewParser()
	err := p.Build(bpoPath)
	if err != nil {
		slog.Warn("Failed to parse bpo.", "error", err)
		return err
	}
	fmt.Printf("Successfully built bpo for %s.\n", p.ModuleName)
	return nil
}
