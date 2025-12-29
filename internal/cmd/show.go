package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yevhenii/git-remarks/internal/git"
	"github.com/yevhenii/git-remarks/internal/store"
)

var showCmd = &cobra.Command{
	Use:   "show [commit]",
	Short: "Show remarks on a specific commit",
	Long: `Show all remarks attached to a specific commit.

If no commit is specified, shows remarks on HEAD.

Examples:
  git remarks show
  git remarks show abc1234`,
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	commit := "HEAD"
	if len(args) > 0 {
		commit = args[0]
	}

	// Resolve commit to full SHA
	fullSHA, err := git.Run("rev-parse", commit)
	if err != nil {
		return fmt.Errorf("invalid commit: %s", commit)
	}

	shortSHA, _ := git.GetShortSHA(fullSHA)

	s := store.New()
	remarks, err := s.Get(fullSHA)
	if err != nil {
		return fmt.Errorf("failed to get remarks: %w", err)
	}

	if remarks.IsEmpty() {
		fmt.Printf("%s — no remarks\n", shortSHA)
		return nil
	}

	isHead := commit == "HEAD"
	headIndicator := ""
	if isHead {
		headIndicator = " (HEAD)"
	}

	// Count active vs all
	activeCount := 0
	for _, r := range remarks.Remarks {
		if r.State == "active" {
			activeCount++
		}
	}

	fmt.Printf("%s%s — %d remark%s\n\n", shortSHA, headIndicator, len(remarks.Remarks), pluralize(len(remarks.Remarks)))

	for _, r := range remarks.Remarks {
		stateIndicator := ""
		if r.State == "resolved" {
			stateIndicator = " [resolved]"
		}

		age := formatAge(r.CreatedAt)
		fmt.Printf("[%s] %s · %s · %s%s\n", r.ID, r.Type, age, r.Branch, stateIndicator)

		// Indent the body
		lines := strings.Split(strings.TrimSpace(r.Body), "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	return nil
}

