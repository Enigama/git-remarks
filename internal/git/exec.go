package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// NotesRef is the git notes reference used for remarks
const NotesRef = "remarks"

// Run executes a git command and returns the output
func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), stderrStr)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunWithStdin executes a git command with stdin input
func RunWithStdin(stdin string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), stderrStr)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsInsideWorkTree checks if the current directory is inside a git repository
func IsInsideWorkTree() bool {
	output, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil && output == "true"
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot() (string, error) {
	return Run("rev-parse", "--show-toplevel")
}

// GetGitDir returns the .git directory path
func GetGitDir() (string, error) {
	return Run("rev-parse", "--git-dir")
}

