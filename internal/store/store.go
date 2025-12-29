package store

import (
	"strings"

	"github.com/Enigama/git-remarks/internal/git"
	"github.com/Enigama/git-remarks/internal/remark"
)

// Store handles reading and writing remarks to git notes
type Store struct {
	notesRef string
}

// New creates a new Store
func New() *Store {
	return &Store{
		notesRef: git.NotesRef,
	}
}

// Get retrieves remarks for a commit
func (s *Store) Get(commit string) (*remark.Remarks, error) {
	output, err := git.Run("notes", "--ref="+s.notesRef, "show", commit)
	if err != nil {
		// No notes found for this commit
		if strings.Contains(strings.ToLower(err.Error()), "no note found") {
			return &remark.Remarks{}, nil
		}
		return nil, err
	}

	return remark.ParseRemarks([]byte(output))
}

// Save writes remarks to a commit, overwriting any existing notes
func (s *Store) Save(commit string, remarks *remark.Remarks) error {
	if remarks.IsEmpty() {
		// Remove notes if no remarks
		return s.Remove(commit)
	}

	data, err := remarks.Marshal()
	if err != nil {
		return err
	}

	_, err = git.RunWithStdin(string(data), "notes", "--ref="+s.notesRef, "add", "-f", "-F", "-", commit)
	return err
}

// Remove removes notes from a commit
func (s *Store) Remove(commit string) error {
	_, err := git.Run("notes", "--ref="+s.notesRef, "remove", commit)
	if err != nil {
		// Ignore error if no notes exist
		if strings.Contains(strings.ToLower(err.Error()), "no note found") {
			return nil
		}
		return err
	}
	return nil
}

// Add adds a new remark to a commit
func (s *Store) Add(commit string, r remark.Remark) error {
	remarks, err := s.Get(commit)
	if err != nil {
		return err
	}

	remarks.Add(r)
	return s.Save(commit, remarks)
}

// Resolve marks a remark as resolved by removing it
func (s *Store) Resolve(commit, remarkID string) (bool, error) {
	remarks, err := s.Get(commit)
	if err != nil {
		return false, err
	}

	if !remarks.RemoveByID(remarkID) {
		return false, nil
	}

	return true, s.Save(commit, remarks)
}

// UpdateRemark updates an existing remark
func (s *Store) UpdateRemark(commit string, r remark.Remark) error {
	remarks, err := s.Get(commit)
	if err != nil {
		return err
	}

	found := remarks.FindByID(r.ID)
	if found == nil {
		return nil
	}

	*found = r
	return s.Save(commit, remarks)
}

// ListAllWithRemarks returns all commits that have remarks
func (s *Store) ListAllWithRemarks() (map[string]*remark.Remarks, error) {
	// List all notes in the remarks ref
	output, err := git.Run("notes", "--ref="+s.notesRef, "list")
	if err != nil {
		// No notes exist
		if strings.Contains(err.Error(), "No notes") {
			return make(map[string]*remark.Remarks), nil
		}
		// Check for unborn branch or no notes
		if strings.Contains(err.Error(), "does not have any notes") {
			return make(map[string]*remark.Remarks), nil
		}
		return nil, err
	}

	if output == "" {
		return make(map[string]*remark.Remarks), nil
	}

	result := make(map[string]*remark.Remarks)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Format: <note-object> <annotated-object>
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			commit := parts[1]
			remarks, err := s.Get(commit)
			if err != nil {
				continue
			}
			if !remarks.IsEmpty() {
				result[commit] = remarks
			}
		}
	}

	return result, nil
}

// Migrate moves remarks from one commit to another
func (s *Store) Migrate(oldCommit, newCommit string) error {
	oldRemarks, err := s.Get(oldCommit)
	if err != nil {
		return err
	}

	if oldRemarks.IsEmpty() {
		return nil
	}

	newRemarks, err := s.Get(newCommit)
	if err != nil {
		return err
	}

	newRemarks.Merge(oldRemarks)

	if err := s.Save(newCommit, newRemarks); err != nil {
		return err
	}

	return s.Remove(oldCommit)
}

// FindRemarkByID searches all remarks to find one by ID
// Returns the commit SHA and the remark if found
func (s *Store) FindRemarkByID(id string) (string, *remark.Remark, error) {
	allRemarks, err := s.ListAllWithRemarks()
	if err != nil {
		return "", nil, err
	}

	for commit, remarks := range allRemarks {
		if r := remarks.FindByID(id); r != nil {
			return commit, r, nil
		}
	}

	return "", nil, nil
}

