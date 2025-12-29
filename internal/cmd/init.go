package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/Enigama/git-remarks/internal/git"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Install git-remarks hooks in the current repository",
	Long: `Install the post-rewrite hook in the current repository.

This hook ensures that remarks survive rebases and amends by
migrating them to the new commit SHAs.

Examples:
  git remarks init`,
	RunE: runInit,
}

const postRewriteHook = `#!/bin/sh
# git-remarks post-rewrite hook
# Migrates remarks when commits are rewritten (rebase, amend)

rewrite_type="$1"

# Check if git-remarks is available
if ! command -v git-remarks >/dev/null 2>&1; then
    exit 0
fi

# Pass stdin to git-remarks migrate-rewrites
git-remarks migrate-rewrites "$rewrite_type"
`

func runInit(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	gitDir, err := git.GetGitDir()
	if err != nil {
		return fmt.Errorf("failed to get git directory: %w", err)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	hookPath := filepath.Join(hooksDir, "post-rewrite")

	// Check if hooks directory exists
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Hook exists - check if it's ours
		content, err := os.ReadFile(hookPath)
		if err != nil {
			return fmt.Errorf("failed to read existing hook: %w", err)
		}

		if containsRemarksHook(string(content)) {
			fmt.Println("✓ git-remarks hook already installed")
			return nil
		}

		// Append to existing hook
		f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			return fmt.Errorf("failed to open existing hook: %w", err)
		}
		defer f.Close()

		appendHook := `
# git-remarks hook (appended)
if command -v git-remarks >/dev/null 2>&1; then
    git-remarks migrate-rewrites "$1"
fi
`
		if _, err := f.WriteString(appendHook); err != nil {
			return fmt.Errorf("failed to append to hook: %w", err)
		}

		fmt.Println("✓ git-remarks hook appended to existing post-rewrite hook")
		return nil
	}

	// Create new hook
	if err := os.WriteFile(hookPath, []byte(postRewriteHook), 0755); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	fmt.Println("✓ git-remarks hook installed")
	return nil
}

func containsRemarksHook(content string) bool {
	return len(content) > 0 && 
		(contains(content, "git-remarks") || contains(content, "git remarks"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

