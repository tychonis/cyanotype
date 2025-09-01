package diff

import (
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare built out from two version",
	Run:   run,
	Args:  cobra.MinimumNArgs(2),
}

func run(cmd *cobra.Command, args []string) {
	commitA := args[0]
	commitB := args[1]
	treeA, _ := getCommitTree(".", commitA)
	treeB, _ := getCommitTree(".", commitB)
	buildOnTree(treeA)
	buildOnTree(treeB)
}

func getCommitTree(repo string, commit string) (*object.Tree, error) {
	r, err := git.PlainOpen(repo)
	if err != nil {
		return nil, err
	}
	c, err := r.CommitObject(plumbing.NewHash(commit))
	if err != nil {
		return nil, err
	}
	return c.Tree()
}

func buildOnTree(t *object.Tree) {
	//TODO
}
