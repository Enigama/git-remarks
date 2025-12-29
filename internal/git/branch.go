package git

import (
	"errors"
	"strings"
)

// ErrDetachedHead is returned when the repository is in detached HEAD state
var ErrDetachedHead = errors.New("not on a branch (detached HEAD)")

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	output, err := Run("symbolic-ref", "--short", "HEAD")
	if err != nil {
		// Check if we're in detached HEAD state
		if strings.Contains(err.Error(), "not a symbolic ref") {
			return "", ErrDetachedHead
		}
		return "", err
	}
	return output, nil
}

// IsDetachedHead returns true if the repository is in detached HEAD state
func IsDetachedHead() bool {
	_, err := GetCurrentBranch()
	return errors.Is(err, ErrDetachedHead)
}

// GetHEAD returns the current HEAD commit SHA
func GetHEAD() (string, error) {
	return Run("rev-parse", "HEAD")
}

// GetShortSHA returns the short SHA for a commit
func GetShortSHA(commit string) (string, error) {
	return Run("rev-parse", "--short", commit)
}

// GetAncestors returns all ancestor commits from the given commit
// up to the specified limit (0 = no limit)
func GetAncestors(commit string, limit int) ([]string, error) {
	args := []string{"rev-list", commit}
	if limit > 0 {
		args = append(args, "-n", string(rune(limit+'0')))
	}

	output, err := Run(args...)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetAncestorsUpTo returns all ancestor commits from start to end (exclusive)
func GetAncestorsUpTo(start, end string) ([]string, error) {
	output, err := Run("rev-list", start, "^"+end)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// IsAncestor returns true if ancestor is an ancestor of descendant
func IsAncestor(ancestor, descendant string) (bool, error) {
	_, err := Run("merge-base", "--is-ancestor", ancestor, descendant)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetCommitLog returns a list of commits with their short SHAs and subjects
func GetCommitLog(commit string, limit int) ([]CommitInfo, error) {
	args := []string{"log", "--format=%H %h %s", commit}
	if limit > 0 {
		args = append(args, "-n", string(rune(limit+'0')))
	}

	output, err := Run(args...)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 2 {
			info := CommitInfo{
				SHA:      parts[0],
				ShortSHA: parts[1],
			}
			if len(parts) == 3 {
				info.Subject = parts[2]
			}
			commits = append(commits, info)
		}
	}

	return commits, nil
}

// CommitInfo contains basic information about a commit
type CommitInfo struct {
	SHA      string
	ShortSHA string
	Subject  string
}

