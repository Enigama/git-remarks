package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Enigama/git-remarks/internal/git"
	"github.com/Enigama/git-remarks/internal/store"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <id>",
	Short: "Resolve (remove) a remark",
	Long: `Mark a remark as resolved by removing it.

The remark is identified by its ID (shown in list/show output).

Examples:
  git remarks resolve a1b2c3d4`,
	Args: cobra.ExactArgs(1),
	RunE: runResolve,
}

func runResolve(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	remarkID := args[0]

	s := store.New()

	// Find the remark
	commit, r, err := s.FindRemarkByID(remarkID)
	if err != nil {
		return fmt.Errorf("failed to find remark: %w", err)
	}

	if r == nil {
		return fmt.Errorf("remark not found: %s", remarkID)
	}

	// Remove the remark
	found, err := s.Resolve(commit, remarkID)
	if err != nil {
		return fmt.Errorf("failed to resolve remark: %w", err)
	}

	if !found {
		return fmt.Errorf("remark not found: %s", remarkID)
	}

	fmt.Printf("✓ Resolved [%s]\n", remarkID)
	return nil
}

