package scene

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "scene",
	Short: "Manage scenes",
}

func init() {
	Cmd.AddCommand(
		GenerateCmd,
		UploadCmd,
		MakeCmd,
	)
}
