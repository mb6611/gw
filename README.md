# gw

A fast git worktree manager with fzf integration, optimized for Claude Code sessions.

## Features

- **fzf-powered switching** — Quickly jump between worktrees with fuzzy search
- **Environment symlinking** — Automatically symlinks `.env`, `.envrc`, `.claude/` to new worktrees
- **Claude integration** — Optional `-c` flag to launch Claude Code after switching
- **Works from any worktree** — Detects siblings automatically

## Installation

### Homebrew

```bash
brew tap mb6611/tap
brew install gw
```

### Go

```bash
go install github.com/mb6611/gw/cmd/gw@latest
```

### Shell integration

Add to your shell config:

```bash
# ~/.zshrc
eval "$(gw init zsh)"

# ~/.bashrc
eval "$(gw init bash)"

# ~/.config/fish/config.fish
gw init fish | source
```

## Usage

### Switch worktrees

```bash
gw              # fzf picker to switch
gw -c           # switch and launch claude
```

### Create new worktree

```bash
gw new feature-auth       # creates ../repo-feature-auth/
gw new feature-auth -c    # create and launch claude
```

New worktrees are created as sibling directories and automatically get symlinks for:
- `.env`, `.env.*`
- `.envrc`
- `.claude/`

### List worktrees

```bash
gw ls
```

Output:
```
* main          /Users/me/code/myapp
  feature-auth  /Users/me/code/myapp-feature-auth
  feature-ui    /Users/me/code/myapp-feature-ui
```

### Remove worktree

```bash
gw rm           # fzf picker to remove
```

## Requirements

- [fzf](https://github.com/junegunn/fzf) — `brew install fzf`
- Git 2.5+ (worktree support)

## License

MIT
