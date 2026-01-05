# gw - Git Worktree CLI

A CLI tool that provides an fzf-powered interface for managing git worktrees, optimized for Claude Code sessions.

## Commands

| Command | Description |
|---------|-------------|
| `gw` | Interactive fzf picker → cd into selected worktree |
| `gw -c` | Same as above + launch `claude` after switching |
| `gw new <branch>` | Create worktree as sibling dir + copy env files + cd into it |
| `gw new <branch> -c` | Same as above + launch `claude` |
| `gw ls` | List all worktrees (non-interactive) |
| `gw rm` | Interactive fzf picker → remove selected worktree |

## Behaviors

### Worktree Detection
- Works from **any worktree** in the repo
- Detects the git root and lists all sibling worktrees
- Example: Running `gw` from `~/code/repo-feature-a/` shows all worktrees including `repo-main/`, `repo-feature-b/`, etc.

### Directory Naming
- New worktrees created as **sibling directories**
- Pattern: `../<repo-name>-<branch-name>/`
- Example: `gw new feature-auth` from `~/code/myapp/` creates `~/code/myapp-feature-auth/`

### Environment File Symlinking
On `gw new`, automatically **symlink** these files/directories from the source worktree to the new one:
- `.env`
- `.env.*` (all env variants)
- `.envrc`
- `.claude/` (entire directory)

Symlinks ensure:
- Single source of truth for env configuration
- Changes propagate automatically across all worktrees
- No drift between worktrees

### Shell Integration
- The binary outputs the target path
- A shell function wraps it to perform the actual `cd`
- Required for the cd-on-switch functionality

### Claude Launch (`-c` flag)
- When `-c` is passed, launch `claude` in the target directory after switching
- Works on: `gw`, `gw new`
- Does not apply to: `gw ls`, `gw rm`

## Technical Details

### Tech Stack
- **Language**: Go
- **Dependencies**: Requires `fzf` to be installed
- **Distribution**: Homebrew formula

### Shell Function (installed via Homebrew or manually)

```bash
# Add to ~/.bashrc or ~/.zshrc
gw() {
  local result
  result=$(command gw "$@")
  local exit_code=$?

  if [[ $exit_code -eq 0 && -d "$result" ]]; then
    cd "$result"
  elif [[ -n "$result" ]]; then
    echo "$result"
  fi

  return $exit_code
}
```

### Exit Codes
- `0`: Success (output is a path to cd into)
- `1`: Error (output is error message)
- `2`: User cancelled (no output)

## Homebrew Distribution

Create a Homebrew tap with:
- Formula that installs the `gw` binary
- Post-install message instructing user to add shell function

## File Structure

```
gw/
├── cmd/
│   └── gw/
│       └── main.go
├── internal/
│   ├── worktree/      # Git worktree operations
│   ├── fzf/           # fzf integration
│   ├── env/           # Environment file symlinking
│   └── claude/        # Claude launcher
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Future Considerations (Out of Scope for v1)
- Config file for customizing copied files
- Worktree cleanup commands
- Integration with other editors/tools
