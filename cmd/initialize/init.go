// initialize a folder for cyanotype
// init is a golang keyword therefore the long name here.
package initialize

import (
	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/internal/catalog"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize current folder",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	catalog.Initialize()
}
