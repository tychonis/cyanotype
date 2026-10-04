package scene

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

var CacheCmd = &cobra.Command{
	Use:   "cache <server> <item> <file>",
	Short: "Cache item representation file",
	RunE:  runCache,
	Args:  cobra.ExactArgs(3),
}

func init() {
	CacheCmd.Flags().StringP("tag", "t", "scene", "set cache tag")
}

func runCache(cmd *cobra.Command, args []string) error {
	server := args[0]
	qualifier := args[1]
	filename := args[2]
	token := os.Getenv("BOMHUB_TOKEN")

	cat := catalog.New("local")
	sym, err := cat.FindCurrent(qualifier)
	if err != nil {
		slog.Error("Failed to find item.", "error", err)
		return err
	}
	sym.GetDigest()

	fileDigest, err := digest.SHA256FromFile(filename)
	if err != nil {
		slog.Error("Failed to compute file digest.", "error", err)
		return err
	}
	artifact := &model.Artifact{
		Filename: path.Base(filename),
		Tag:      cmd.Flag("tag").Value.String(),
		Source:   fmt.Sprintf("file://%s", filename),
		Digest:   fileDigest,
	}

	return RegisterRepresentation(server, token, sym.GetDigest(), artifact)
}

func RegisterRepresentation(server string, token string, sym model.Digest, artifact *model.Artifact) error {
	client := catalog.NewHTTPClient(token)
	storage := catalog.NewAPIStore(server, client)
	err := storage.SaveArtifact(artifact)
	if err != nil {
		return err
	}

	repEndpoint := fmt.Sprintf("%s/representation/%s", server, sym)
	payload, err := json.Marshal(artifact)
	if err != nil {
		return err
	}
	resp, err := client.Post(repEndpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to register representation: %s", resp.Status)
	}
	return nil
}
