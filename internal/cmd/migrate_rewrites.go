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

var migrateRewritesCmd = &cobra.Command{
	Use:    "migrate-rewrites [rewrite-type]",
	Short:  "Migrate remarks after a rewrite (internal command)",
	Hidden: true,
	RunE:   runMigrateRewrites,
}

func runMigrateRewrites(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return nil // Silently exit if not in a git repo
	}

	s := store.New()

	// Read old-sha new-sha pairs from stdin
	scanner := bufio.NewScanner(os.Stdin)
	migratedCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		oldSHA := parts[0]
		newSHA := parts[1]

		// Check if old commit has remarks
		oldRemarks, err := s.Get(oldSHA)
		if err != nil {
			continue
		}

		if oldRemarks.IsEmpty() {
			continue
		}

		// Migrate remarks to new commit
		if err := s.Migrate(oldSHA, newSHA); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to migrate remarks from %s to %s: %v\n", oldSHA[:7], newSHA[:7], err)
			continue
		}

		migratedCount++
		shortOld := oldSHA
		shortNew := newSHA
		if len(oldSHA) > 7 {
			shortOld = oldSHA[:7]
		}
		if len(newSHA) > 7 {
			shortNew = newSHA[:7]
		}
		fmt.Printf("Migrated %d remark(s): %s → %s\n", len(oldRemarks.Remarks), shortOld, shortNew)
	}

	return nil
}

