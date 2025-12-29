package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yevhenii/git-remarks/internal/git"
	"github.com/yevhenii/git-remarks/internal/remark"
	"github.com/yevhenii/git-remarks/internal/store"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit an existing remark",
	Long: `Edit an existing remark in your $EDITOR.

The remark is identified by its ID (shown in list/show output).

Examples:
  git remarks edit a1b2c3d4`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	remarkID := args[0]

	s := store.New()

	// Find the remark
	commit, r, err := s.FindRemarkByID(remarkID)
	if err != nil {
		return fmt.Errorf("failed to find remark: %w", err)
	}

	if r == nil {
		return fmt.Errorf("remark not found: %s", remarkID)
	}

	shortSHA, _ := git.GetShortSHA(commit)

	// Open editor with current content
	newBody, newType, err := openEditorForEdit(shortSHA, r)
	if err != nil {
		return err
	}

	if strings.TrimSpace(newBody) == "" {
		return fmt.Errorf("remark body cannot be empty")
	}

	// Update the remark
	r.Body = newBody
	if newType != "" && remark.ValidateType(newType) {
		r.Type = remark.Type(newType)
	}

	if err := s.UpdateRemark(commit, *r); err != nil {
		return fmt.Errorf("failed to update remark: %w", err)
	}

	fmt.Printf("✓ Updated [%s]\n", remarkID)
	return nil
}

func openEditorForEdit(commit string, r *remark.Remark) (body string, newType string, err error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Create temp file with current content
	tmpfile, err := os.CreateTemp("", "git-remark-*.yaml")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	template := fmt.Sprintf(`# Editing remark [%s] on %s (%s)
# Lines starting with # are ignored

type: %s

---

%s`, r.ID, commit, r.Branch, r.Type, r.Body)

	if _, err := tmpfile.WriteString(template); err != nil {
		return "", "", fmt.Errorf("failed to write template: %w", err)
	}
	tmpfile.Close()

	// Open editor
	editorCmd := exec.Command(editor, tmpfile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", "", fmt.Errorf("editor failed: %w", err)
	}

	// Read back the file
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return parseEditorContent(string(content))
}

