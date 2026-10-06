package scene

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var RootCmd = &cobra.Command{
	Use:   "root <server> <item>",
	Short: "Add root scene",
	RunE:  runRoot,
	Args:  cobra.ExactArgs(2),
}

func runRoot(cmd *cobra.Command, args []string) error {
	server := args[0]
	qualifier := args[1]

	token := os.Getenv("BOMHUB_TOKEN")

	localCat := catalog.New("local")
	remoteCat := catalog.NewRemoteCatalog(server, token, localCat.ModuleName)
	return remoteCat.AddRoot(server, token, qualifier)
}
