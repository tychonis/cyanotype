package symbol

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "symbol",
	Short: "Manage symbols",
}

func init() {
	Cmd.AddCommand(
		ListCmd,
	)
}
