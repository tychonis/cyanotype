package pull

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var Cmd = &cobra.Command{
	Use:   "pull <server> [<module>]",
	Short: "Adhoc implementation pulling catalog from remote",
	RunE:  run,
	Args:  cobra.RangeArgs(1, 2),
}

func run(cmd *cobra.Command, args []string) error {
	server := args[0]
	token := os.Getenv("BOMHUB_TOKEN")

	localCat := catalog.New("local")
	if len(args) == 2 {
		module := args[1]
		err := localCat.SetModuleName(module)
		if err != nil {
			return err
		}
	}
	remoteCat := catalog.NewRemoteCatalog(server, token, localCat.ModuleName)
	err := localCat.Pull(remoteCat)
	if err != nil {
		if err == catalog.ErrNoNewRevision {
			slog.Info("No new revisions to pull.")
			return nil
		}
		slog.Error("Failed to pull catalog from remote.", "error", err)
		return err
	}
	return nil
}
