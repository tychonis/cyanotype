package plan

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "plan",
	Short: "plan shows the diff between working-tree and committed state",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {}
