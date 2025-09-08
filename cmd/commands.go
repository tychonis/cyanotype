package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/cmd/bom"
	"github.com/tychonis/cyanotype/cmd/build"
	"github.com/tychonis/cyanotype/cmd/commit"
	"github.com/tychonis/cyanotype/cmd/initialize"
)

var rootCmd = &cobra.Command{
	Use:   "cyanotype",
	Short: "cyanotype manages bom as code",
	Long:  "TODO: Add doc string",
}

func Run() {
	rootCmd.AddCommand(
		initialize.Cmd,
		bom.Cmd,
		build.Cmd,
		commit.Cmd,
	)

	rootCmd.Execute()
}
