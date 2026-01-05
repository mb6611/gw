# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build -o gw ./cmd/gw    # Build binary
go run ./cmd/gw            # Run directly
go test ./...              # Run all tests
```

## Architecture

This is a Go CLI tool using cobra for command handling. The binary outputs paths to stdout which a shell wrapper function uses to `cd` into the selected worktree.

**Key architectural pattern**: The Go binary cannot change the parent shell's directory. It outputs paths/markers to stdout, and the shell function (from `gw init <shell>`) interprets that output to perform the actual `cd`.

### Package Structure

- `cmd/gw/main.go` - Entry point, calls `cmd.Execute()`
- `internal/cmd/` - Cobra command definitions (root, switch, new, ls, rm, init)
- `internal/worktree/` - Git worktree operations via `git worktree` CLI
- `internal/fzf/` - fzf picker integration
- `internal/env/` - Environment file symlinking logic

### Shell Integration

The `init` command outputs shell functions that wrap the binary. Critical details:
- Uses `\builtin` and `\command` prefixes to avoid alias conflicts (zoxide pattern)
- Variable names must avoid zsh special variables (`path` is tied to `PATH`)
- Extracts first line of output as the path, checks for markers like `__GW_LAUNCH_CLAUDE__`

### Output Protocol

Commands output a path on the first line. Optional markers on subsequent lines:
- `__GW_LAUNCH_CLAUDE__` - Shell wrapper should run `claude`
- `__GW_LAUNCH_CLAUDE_DANGEROUS__` - Shell wrapper should run `claude --dangerously-skip-permissions`

## Homebrew Distribution

- Formula in `Formula/gw.rb` (also mirrored to `mb6611/homebrew-gw` tap repo)
- Update both when releasing: change version URL and sha256
- Get sha256: `curl -sL <tarball-url> | shasum -a 256`
