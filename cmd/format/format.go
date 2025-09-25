package format

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "format",
	Short: "Format bpo file.",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
}
