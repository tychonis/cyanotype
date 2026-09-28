package push

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "push <server>",
	Short: "Adhoc implementation saving catalog to remote",
	RunE:  run,
	Args:  cobra.ExactArgs(1),
}

func run(cmd *cobra.Command, args []string) error {
	server := args[0]
	token := os.Getenv("BOMHUB_TOKEN")

	localCat := catalog.New("local")
	remoteCat := catalog.NewRemoteCatalog(server, token, localCat.ModuleName)
	err := localCat.Push(remoteCat)
	if err != nil {
		if err == catalog.ErrNoNewRevision {
			slog.Info("No new revisions to push.")
			return nil
		}
		slog.Error("Failed to push catalog to remote.", "error", err)
		return err
	}
	return remoteCat.Upload(server, token)
}
