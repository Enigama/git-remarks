package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yevhenii/git-remarks/internal/git"
	"github.com/yevhenii/git-remarks/internal/store"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover orphaned remarks using patch-id matching",
	Long: `Attempt to recover remarks that are attached to commits no longer in the current branch.

This uses patch-id matching to find commits with the same content
and prompts you to migrate remarks to the new commits.

Examples:
  git remarks recover`,
	RunE: runRecover,
}

func runRecover(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	head, err := git.GetHEAD()
	if err != nil {
		return fmt.Errorf("cannot get HEAD: %w", err)
	}

	s := store.New()
	allRemarks, err := s.ListAllWithRemarks()
	if err != nil {
		return fmt.Errorf("failed to list remarks: %w", err)
	}

	// Find orphaned remarks (commits not in current branch)
	orphaned := make(map[string]struct{})
	for commit := range allRemarks {
		isAncestor, err := git.IsAncestor(commit, head)
		if err != nil || !isAncestor {
			orphaned[commit] = struct{}{}
		}
	}

	if len(orphaned) == 0 {
		fmt.Println("No orphaned remarks found")
		return nil
	}

	fmt.Printf("Found %d orphaned commit(s) with remarks\n\n", len(orphaned))

	reader := bufio.NewReader(os.Stdin)
	recoveredCount := 0

	for oldCommit := range orphaned {
		remarks := allRemarks[oldCommit]
		shortOld, _ := git.GetShortSHA(oldCommit)

		// Get patch-id for the orphaned commit
		patchID, err := git.GetPatchID(oldCommit)
		if err != nil || patchID == "" {
			fmt.Printf("Cannot compute patch-id for %s, skipping\n", shortOld)
			continue
		}

		// Search for matching patch-id in current branch
		newCommit, err := git.FindCommitByPatchID(patchID, head, 1000)
		if err != nil || newCommit == "" {
			fmt.Printf("No matching commit found for %s, skipping\n", shortOld)
			continue
		}

		shortNew, _ := git.GetShortSHA(newCommit)

		// Show what we found
		fmt.Printf("Found match: %s → %s\n", shortOld, shortNew)
		for _, r := range remarks.Remarks {
			bodyPreview := r.Body
			if len(bodyPreview) > 50 {
				bodyPreview = bodyPreview[:50] + "..."
			}
			bodyPreview = strings.ReplaceAll(bodyPreview, "\n", " ")
			fmt.Printf("  [%s] %s: %s\n", r.ID, r.Type, bodyPreview)
		}

		// Prompt user
		fmt.Printf("Migrate these remarks? [y/N] ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			if err := s.Migrate(oldCommit, newCommit); err != nil {
				fmt.Printf("Error migrating: %v\n", err)
				continue
			}
			fmt.Printf("✓ Migrated %d remark(s)\n", len(remarks.Remarks))
			recoveredCount++
		} else {
			fmt.Println("Skipped")
		}
		fmt.Println()
	}

	if recoveredCount > 0 {
		fmt.Printf("Recovered remarks from %d commit(s)\n", recoveredCount)
	}

	return nil
}

