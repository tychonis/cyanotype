package artifact

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

var RedactCmd = &cobra.Command{
	Use:   "redact",
	Short: "Redact all artifacts",
	RunE:  redactCmd,
}

func init() {
	RedactCmd.Flags().String("backup", "./artifact_backup", "directory for original artifacts")
}

func redactArtifact(art *model.Artifact, backup string) error {
	parsed, err := url.Parse(art.Source)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "file":
		// continue with file handling
		src := parsed.Host + parsed.Path
		dst := filepath.Join(backup, art.Digest)

		h, err := digest.SHA256FromFile(src)
		if err != nil {
			return err
		}
		if h != art.Digest {
			slog.Debug("Artifact digest mismatch, skipping",
				"path", src,
				"expected", art.Digest,
				"actual", h)
			return nil
		}

		if _, err := os.Stat(dst); err == nil {
			h, err := digest.SHA256FromFile(dst)
			if err != nil {
				return err
			}
			if h != art.Digest {
				return fmt.Errorf("backup digest mismatch: %s", dst)
			}
		} else if os.IsNotExist(err) {
			if err := os.Rename(src, dst); err != nil {
				return err
			}
		} else {
			return err
		}

		content := []byte("redacted content with digest: " + art.Digest + "\n")
		return os.WriteFile(src, content, 0644)
	default:
		return fmt.Errorf("unsupported artifact source scheme: %s", parsed.Scheme)
	}
}

func redactCmd(cmd *cobra.Command, args []string) error {
	backup, err := cmd.Flags().GetString("backup")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(backup, 0755); err != nil {
		return err
	}

	cat := catalog.New("local")
	artifacts, err := cat.ListArtifacts()
	if err != nil {
		return err
	}

	for _, artifact := range artifacts {
		err := redactArtifact(artifact, backup)
		if err != nil {
			continue
		}
	}
	return nil
}
