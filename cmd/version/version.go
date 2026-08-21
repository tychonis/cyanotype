package version

import (
	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/version"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	cmd.Println("cyanotype version:", version.Version)
}
