package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/cmd/bom"
	"github.com/tychonis/cyanotype/cmd/build"
)

var rootCmd = &cobra.Command{
	Use:   "cyanotype",
	Short: "cyanotype manages bom as code",
	Long:  "TODO: Add doc string",
}

func Run() {
	rootCmd.AddCommand(
		bom.Cmd,
		build.Cmd,
	)

	rootCmd.Execute()
}
