package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/Enigama/git-remarks/internal/git"
	"github.com/Enigama/git-remarks/internal/store"
)

var migrateBranchCmd = &cobra.Command{
	Use:   "migrate-branch <old-name> <new-name>",
	Short: "Update branch name in all remarks",
	Long: `Rename a branch in all existing remarks.

Use this after renaming a branch to update remarks that reference the old name.

Examples:
  git remarks migrate-branch old-feature new-feature`,
	Args: cobra.ExactArgs(2),
	RunE: runMigrateBranch,
}

func runMigrateBranch(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	oldBranch := args[0]
	newBranch := args[1]

	s := store.New()
	allRemarks, err := s.ListAllWithRemarks()
	if err != nil {
		return fmt.Errorf("failed to list remarks: %w", err)
	}

	updatedCount := 0

	for commit, remarks := range allRemarks {
		needsUpdate := false

		for i := range remarks.Remarks {
			if remarks.Remarks[i].Branch == oldBranch {
				remarks.Remarks[i].Branch = newBranch
				needsUpdate = true
				updatedCount++
			}
		}

		if needsUpdate {
			if err := s.Save(commit, remarks); err != nil {
				return fmt.Errorf("failed to update remarks on %s: %w", commit[:7], err)
			}
		}
	}

	if updatedCount == 0 {
		fmt.Printf("No remarks found for branch '%s'\n", oldBranch)
		return nil
	}

	fmt.Printf("✓ Updated %d remark%s from '%s' to '%s'\n", updatedCount, pluralize(updatedCount), oldBranch, newBranch)
	return nil
}

