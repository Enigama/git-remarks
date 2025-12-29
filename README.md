# git-remarks

Personal developer notes attached to Git commits.

**git-remarks** is a tool for managing personal notes (thoughts, doubts, todos, decisions) attached to Git commits. Notes are scoped to branches and survive rebases. They are stored using `git notes` and remain local by default.

## Features

- **Branch-scoped notes** — Notes are tied to a specific branch and only show when on that branch
- **Survives rebases** — Automatic migration of notes when commits are rewritten
- **Local-first** — Notes stay local, optional to share
- **Simple CLI** — Intuitive commands for common workflows
- **Editor integration** — Edit notes in your `$EDITOR`

## Installation

### From GitHub Releases

Download the latest release from [GitHub Releases](https://github.com/yevhenii/git-remarks/releases) and add the binary to your PATH.

### From Source

```bash
go install github.com/yevhenii/git-remarks/cmd/git-remarks@latest
```

## Quick Start

```bash
# Initialize hooks in your repository
git remarks init

# Add a remark to the current commit
git remarks add "This is a test helper, remove before PR"

# Add a typed remark
git remarks add --type todo "Refactor this later"

# List all active remarks on the current branch
git remarks

# Show remarks on a specific commit
git remarks show abc1234

# Resolve (remove) a remark
git remarks resolve a1b2c3d4
```

## Commands

### `git remarks` / `git remarks list`

List all active remarks on the current branch.

```bash
$ git remarks

feature/auth (2 active remarks)

[a1b2c3d4] thought · 2h ago · abc1234 (HEAD)
  Not confident in this approach yet.

[b2c3d4e5] todo · 1d ago · def5678
  Check if session timeout is handled.
```

### `git remarks add [commit] [body]`

Add a new remark. Opens `$EDITOR` if no body provided.

```bash
# Add to HEAD
git remarks add "Quick note"

# Add with specific type
git remarks add --type doubt "Is this the right abstraction?"

# Add to specific commit
git remarks add abc1234 "Note on older commit"

# Open editor
git remarks add
```

**Types:** `thought` (default), `doubt`, `todo`, `decision`

### `git remarks show [commit]`

Show all remarks on a specific commit (default: HEAD).

### `git remarks resolve <id>`

Mark a remark as resolved (removes it).

### `git remarks edit <id>`

Edit an existing remark in your `$EDITOR`.

### `git remarks init`

Install the post-rewrite hook in the current repository. This ensures remarks survive rebases.

### `git remarks recover`

Recover orphaned remarks using patch-id matching. Useful if hooks weren't installed during a rebase.

### `git remarks migrate-branch <old> <new>`

Update branch name in all remarks after renaming a branch.

## How It Works

### Storage

Remarks are stored using Git's built-in notes system at `refs/notes/remarks`. Each commit can have multiple remarks stored as a YAML document.

### Rebase Survival

When you run `git remarks init`, a `post-rewrite` hook is installed. This hook automatically migrates remarks to new commit SHAs after:

- `git rebase`
- `git commit --amend`
- Interactive rebases with squash/fixup

### Recovery

If remarks become orphaned (e.g., hooks weren't installed), use `git remarks recover` to find matching commits by patch-id and migrate remarks.

## Note Format

Remarks are stored as YAML:

```yaml
remarks:
  - id: a1b2c3d4
    type: thought
    branch: feature/auth
    state: active
    created_at: 2025-12-27T10:30:00Z
    body: |
      Not confident in this approach yet.
      Revisit after more usage.
```

## Philosophy

git-remarks is designed for **personal thinking**, not collaboration:

- Notes represent unresolved thoughts that follow a branch
- They help you track mental context across development sessions
- Git history becomes a mental timeline
- Notes do not modify commits — SHAs remain unchanged

## License

MIT

