package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/cmd/bom"
	"github.com/tychonis/cyanotype/cmd/build"
	"github.com/tychonis/cyanotype/cmd/commit"
	"github.com/tychonis/cyanotype/cmd/export"
	"github.com/tychonis/cyanotype/cmd/initialize"
	"github.com/tychonis/cyanotype/cmd/push"
	"github.com/tychonis/cyanotype/cmd/tree"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "cyanotype",
	Short: "cyanotype manages bom as code",
	Long:  "TODO: Add doc string",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			slog.Debug("Debug logging enabled")
		}
	},
}

func Run() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	rootCmd.AddCommand(
		initialize.Cmd,
		bom.Cmd,
		build.Cmd,
		commit.Cmd,
		export.Cmd,
		tree.Cmd,
		push.Cmd,
	)

	rootCmd.Execute()
}
