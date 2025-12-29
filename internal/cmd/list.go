package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yevhenii/git-remarks/internal/git"
	"github.com/yevhenii/git-remarks/internal/remark"
	"github.com/yevhenii/git-remarks/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active remarks on the current branch",
	Long: `List all active remarks that are relevant to the current branch.

This scans the commit history from HEAD and shows all active remarks
that are scoped to the current branch.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	branch, err := git.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("cannot determine current branch: %w", err)
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

	// Collect active remarks for current branch that are ancestors of HEAD
	type remarkWithCommit struct {
		Commit    string
		ShortSHA  string
		Remark    remark.Remark
		IsHead    bool
		Ancestors int // position in history (0 = HEAD)
	}

	var activeRemarks []remarkWithCommit

	for commit, remarks := range allRemarks {
		// Check if this commit is an ancestor of HEAD (or is HEAD)
		isAncestor := commit == head
		if !isAncestor {
			isAnc, err := git.IsAncestor(commit, head)
			if err != nil {
				continue
			}
			isAncestor = isAnc
		}

		if !isAncestor {
			continue
		}

		active := remarks.ActiveForBranch(branch)
		for _, r := range active {
			shortSHA, _ := git.GetShortSHA(commit)
			activeRemarks = append(activeRemarks, remarkWithCommit{
				Commit:   commit,
				ShortSHA: shortSHA,
				Remark:   r,
				IsHead:   commit == head,
			})
		}
	}

	if len(activeRemarks) == 0 {
		fmt.Printf("%s (no active remarks)\n", branch)
		return nil
	}

	fmt.Printf("%s (%d active remark%s)\n\n", branch, len(activeRemarks), pluralize(len(activeRemarks)))

	for _, r := range activeRemarks {
		headIndicator := ""
		if r.IsHead {
			headIndicator = " (HEAD)"
		}

		age := formatAge(r.Remark.CreatedAt)
		fmt.Printf("[%s] %s · %s · %s%s\n", r.Remark.ID, r.Remark.Type, age, r.ShortSHA, headIndicator)
		
		// Indent the body
		lines := strings.Split(strings.TrimSpace(r.Remark.Body), "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	return nil
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func formatAge(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

