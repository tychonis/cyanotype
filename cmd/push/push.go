package push

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "push <path> <server> <tag>",
	Short: "Adhoc implementation saving catalog to remote",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	bpoPath := args[0]
	if bpoPath == "" {
		bpoPath = "."
	}

	server := args[1]
	tag := args[2]
	token := os.Getenv("BOMHUB_TOKEN")

	localCat := catalog.New("local")
	remoteCat := catalog.NewRemoteCatalog(server, token, tag)
	localCat.Push(remoteCat)
}
