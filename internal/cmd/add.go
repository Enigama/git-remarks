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

var (
	addType   string
	addBranch string
	addEdit   bool
)

var addCmd = &cobra.Command{
	Use:   "add [commit] [body]",
	Short: "Add a remark to a commit",
	Long: `Add a new remark to a commit.

If no commit is specified, the remark is added to HEAD.
If no body is specified, your $EDITOR will open.

Examples:
  git remarks add "This is a test helper, remove before PR"
  git remarks add --type todo "Refactor this later"
  git remarks add abc1234 "Note on older commit"
  git remarks add  # opens editor`,
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addType, "type", "t", "thought", "Remark type: thought, doubt, todo, decision")
	addCmd.Flags().StringVarP(&addBranch, "branch", "b", "", "Override branch (for detached HEAD)")
	addCmd.Flags().BoolVarP(&addEdit, "edit", "e", false, "Force open editor even if body provided")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	// Validate type
	if !remark.ValidateType(addType) {
		return fmt.Errorf("invalid type: %s (must be thought, doubt, todo, or decision)", addType)
	}

	// Determine branch
	branch := addBranch
	if branch == "" {
		var err error
		branch, err = git.GetCurrentBranch()
		if err != nil {
			return fmt.Errorf("not on a branch. Use --branch to specify: git remarks add --branch <name> \"body\"")
		}
	}

	// Determine commit and body from args
	commit := "HEAD"
	var body string

	switch len(args) {
	case 0:
		// No args - will open editor
	case 1:
		// Could be commit or body
		// If it looks like a commit SHA, treat as commit
		if looksLikeCommit(args[0]) {
			commit = args[0]
		} else {
			body = args[0]
		}
	case 2:
		commit = args[0]
		body = args[1]
	default:
		// Join remaining args as body
		commit = args[0]
		body = strings.Join(args[1:], " ")
	}

	// Resolve commit to full SHA
	fullSHA, err := git.Run("rev-parse", commit)
	if err != nil {
		return fmt.Errorf("invalid commit: %s", commit)
	}

	shortSHA, _ := git.GetShortSHA(fullSHA)

	// Open editor if no body or --edit flag
	if body == "" || addEdit {
		editedBody, editedType, err := openEditor(shortSHA, branch, addType, body)
		if err != nil {
			return err
		}
		body = editedBody
		if editedType != "" {
			addType = editedType
		}
	}

	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("remark body cannot be empty")
	}

	// Create and save remark
	r := remark.NewRemark(remark.Type(addType), branch, body)

	s := store.New()
	if err := s.Add(fullSHA, r); err != nil {
		return fmt.Errorf("failed to add remark: %w", err)
	}

	fmt.Printf("✓ Added remark [%s] to %s (%s)\n", r.ID, shortSHA, r.Type)
	return nil
}

func looksLikeCommit(s string) bool {
	// Check if it's a valid git ref
	_, err := git.Run("rev-parse", "--verify", s+"^{commit}")
	return err == nil
}

func openEditor(commit, branch, remarkType, existingBody string) (body string, newType string, err error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Create temp file with template
	tmpfile, err := os.CreateTemp("", "git-remark-*.yaml")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	template := fmt.Sprintf(`# New remark on %s (%s)
# Lines starting with # are ignored

type: %s

---

%s`, commit, branch, remarkType, existingBody)

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

func parseEditorContent(content string) (body string, remarkType string, err error) {
	lines := strings.Split(content, "\n")
	
	var bodyLines []string
	inBody := false
	remarkType = ""

	for _, line := range lines {
		if inBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Check for type line
		if strings.HasPrefix(strings.TrimSpace(line), "type:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				t := strings.TrimSpace(parts[1])
				if remark.ValidateType(t) {
					remarkType = t
				}
			}
			continue
		}

		// Check for separator
		if strings.TrimSpace(line) == "---" {
			inBody = true
			continue
		}
	}

	body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return body, remarkType, nil
}

