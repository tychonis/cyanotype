package scene

import (
	_ "embed"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var MakeCmd = &cobra.Command{
	Use:   "make <server> <item>",
	Short: "Generate makefile managing scenes.",
	RunE:  runMake,
	Args:  cobra.ExactArgs(2),
}

//go:embed assets/Makefile
var makefileTemplate string

func runMake(cmd *cobra.Command, args []string) error {
	server := args[0]
	item := args[1]

	makefileContent := makefileTemplate
	makefileContent = strings.ReplaceAll(makefileContent, "{ROOT}", item)
	makefileContent = strings.ReplaceAll(makefileContent, "{SERVER}", server)

	output := "Makefile"

	return os.WriteFile(output, []byte(makefileContent), 0644)
}
