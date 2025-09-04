package initialize

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "initialize current folder",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	bpcDir := ".bpc"
	stat, err := os.Stat(bpcDir)
	if err == nil {
		if !stat.IsDir() {
			slog.Error("invalid .bpc format")
		}
		return
	}

	err = os.Mkdir(bpcDir, 0755)
	if err != nil {
		return
	}

	fmt.Println("Initialized empty cyanotype repo in .bpc/")
}
