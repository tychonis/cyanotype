package artifact

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "artifact",
	Short: "Manage artifacts",
}

func init() {
	Cmd.AddCommand(ListCmd)
	Cmd.AddCommand(RedactCmd)
	Cmd.AddCommand(RestoreCmd)
}
