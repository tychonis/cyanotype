package artifact

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tychonis/cyanotype/core/catalog"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all artifacts",
	RunE:  listCmd,
}

func listCmd(cmd *cobra.Command, args []string) error {
	cat := catalog.New("local")
	artifacts, err := cat.ListArtifacts()
	if err != nil {
		return err
	}
	for _, artifact := range artifacts {
		fmt.Println(artifact.Source, artifact.Digest)
	}
	return nil
}
