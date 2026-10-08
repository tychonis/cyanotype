package artifact

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/model"
)

var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore all artifacts from backup",
	RunE:  restoreCmd,
}

func init() {
	RestoreCmd.Flags().String("backup", "./artifact_backup", "directory for original artifacts")
}

func restoreArtifact(art *model.Artifact, backup string) error {
	if !strings.HasPrefix(art.Source, "file://") {
		return nil
	}

	src := filepath.Join(backup, art.Digest)
	dst := strings.TrimPrefix(art.Source, "file://")

	h, err := digest.SHA256FromFile(src)
	if err != nil {
		return err
	}
	if h != art.Digest {
		slog.Warn("Backup digest mismatch, skipping",
			"path", src,
			"expected", art.Digest,
			"actual", h)
		return nil
	}

	if h, err := digest.SHA256FromFile(dst); err == nil {
		if h == art.Digest {
			return nil // Already restored
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Copy rather than move: multiple artifacts may share the same digest.
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	defer output.Close()

	info, err := input.Stat()
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode().Perm())
}

func restoreCmd(cmd *cobra.Command, args []string) error {
	backup, err := cmd.Flags().GetString("backup")
	if err != nil {
		return err
	}

	cat := catalog.New("local")
	artifacts, err := cat.ListArtifacts()
	if err != nil {
		return err
	}

	for _, artifact := range artifacts {
		if err := restoreArtifact(artifact, backup); err != nil {
			return err
		}
	}
	return nil
}
