// initialize a folder for cyanotype
// init is a golang keyword therefore the long name here.
package initialize

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize current folder",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	err := catalog.Initialize()
	if err != nil {
		fmt.Println("Failed to initialize:", err)
		return
	}
	fmt.Println("Initialized empty cyanotype repo in .bpc/")
}
