package git

import (
	"strings"
)

// GetPatchID computes the patch-id for a commit
func GetPatchID(commit string) (string, error) {
	// Get the diff for the commit
	diff, err := Run("diff-tree", "-p", commit)
	if err != nil {
		return "", err
	}

	if diff == "" {
		return "", nil
	}

	// Compute patch-id from the diff
	output, err := RunWithStdin(diff, "patch-id", "--stable")
	if err != nil {
		return "", err
	}

	// patch-id output format: <patch-id> <commit-id>
	parts := strings.Fields(output)
	if len(parts) >= 1 {
		return parts[0], nil
	}

	return "", nil
}

// FindCommitByPatchID searches for a commit with matching patch-id in the given range
func FindCommitByPatchID(patchID, startCommit string, limit int) (string, error) {
	if patchID == "" {
		return "", nil
	}

	commits, err := GetAncestors(startCommit, limit)
	if err != nil {
		return "", err
	}

	for _, commit := range commits {
		commitPatchID, err := GetPatchID(commit)
		if err != nil {
			continue
		}
		if commitPatchID == patchID {
			return commit, nil
		}
	}

	return "", nil
}

