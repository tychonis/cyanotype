package push

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "push <server> <tag>",
	Short: "Adhoc implementation saving catalog to remote",
	RunE:  run,
	Args:  cobra.ExactArgs(2),
}

func run(cmd *cobra.Command, args []string) error {
	server := args[0]
	tag := args[1]
	token := os.Getenv("BOMHUB_TOKEN")

	localCat := catalog.New("local")
	remoteCat := catalog.NewRemoteCatalog(server, token, tag)
	err := localCat.Push(remoteCat)
	if err != nil {
		slog.Error("Failed to push catalog to remote.", "error", err)
		return err
	}
	remoteCat.Upload(server, token, tag)
	return nil
}
