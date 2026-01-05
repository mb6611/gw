# Implementation Plan: gw CLI

## Overview

Build a Go CLI tool that provides fzf-powered git worktree management, optimized for Claude Code sessions. The tool outputs paths for shell integration to handle `cd`, and optionally launches `claude`.

## Phases

1. **Project Scaffolding** - Go module, directory structure, Makefile
2. **Core Worktree Operations** - Git worktree list/add/remove via exec
3. **FZF Integration** - Interactive picker piped to fzf
4. **Environment Symlinking** - Symlink .env*, .envrc, .claude/ on new
5. **CLI Commands** - Wire up commands with cobra
6. **Homebrew Distribution** - Tap and formula

---

## Phase 1: Project Scaffolding

### Files to Create

**go.mod**
```
module github.com/1unoe/gw

go 1.22
```

**cmd/gw/main.go** (stub)
```go
package main

import "github.com/1unoe/gw/internal/cmd"

func main() {
    cmd.Execute()
}
```

**internal/cmd/root.go** (stub)
```go
package cmd

import (
    "os"
    "github.com/spf13/cobra"
)

var claudeFlag bool

var rootCmd = &cobra.Command{
    Use:   "gw",
    Short: "Git worktree manager with fzf",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().BoolVarP(&claudeFlag, "claude", "c", false, "Launch claude after switching")
}
```

**Makefile**
```makefile
.PHONY: build install clean

BINARY=gw
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/gw

install: build
	mv $(BINARY) /usr/local/bin/

clean:
	rm -f $(BINARY)
```

### Verification
- `go mod tidy` succeeds
- `make build` produces `gw` binary
- `./gw --help` shows usage

---

## Phase 2: Core Worktree Operations

### internal/worktree/worktree.go

```go
package worktree

type Worktree struct {
    Path   string // Absolute path to worktree
    Branch string // Branch name (or HEAD commit if detached)
    Bare   bool   // Is this the bare repo
}

// List returns all worktrees for the repo containing the current directory
func List() ([]Worktree, error)

// Add creates a new worktree at the given path for the given branch
// If branch doesn't exist, creates it from current HEAD
func Add(path, branch string) error

// Remove deletes a worktree (must not be current directory)
func Remove(path string) error

// Current returns the worktree for the current directory
func Current() (*Worktree, error)

// RepoName returns the base repository name (used for sibling naming)
func RepoName() (string, error)
```

### Implementation Details

**List()** - Parse `git worktree list --porcelain`:
```
worktree /path/to/main
HEAD abc123
branch refs/heads/main

worktree /path/to/feature
HEAD def456
branch refs/heads/feature
```

**Add()** - Execute `git worktree add <path> -b <branch>` or `git worktree add <path> <branch>` if branch exists

**Remove()** - Execute `git worktree remove <path>`

**Current()** - Get current dir, match against List() results

**RepoName()** - Extract from `git rev-parse --show-toplevel`, take basename, strip any existing branch suffix pattern

### Sibling Path Generation

```go
// SiblingPath generates the path for a new worktree
// From /code/myapp or /code/myapp-main, creating branch "feature":
// Returns /code/myapp-feature
func SiblingPath(branch string) (string, error) {
    repoName, _ := RepoName()
    current, _ := Current()
    parent := filepath.Dir(current.Path)
    return filepath.Join(parent, repoName+"-"+branch), nil
}
```

### Verification
- From any worktree, `List()` returns all siblings
- `Add()` creates worktree in correct sibling location
- `Remove()` cleans up worktree

---

## Phase 3: FZF Integration

### internal/fzf/fzf.go

```go
package fzf

// Pick shows worktrees in fzf and returns the selected path
// Returns ("", ErrCancelled) if user presses ESC
// Returns ("", err) on fzf execution error
func Pick(worktrees []worktree.Worktree, prompt string) (string, error)
```

### Implementation

1. Check `fzf` is in PATH, return clear error if not
2. Format worktrees for display: `branch-name  →  /path/to/worktree`
3. Pipe to fzf via exec:
   ```go
   cmd := exec.Command("fzf", "--prompt", prompt, "--height", "40%", "--reverse")
   cmd.Stdin = strings.NewReader(formattedList)
   cmd.Stderr = os.Stderr
   output, err := cmd.Output()
   ```
4. Parse selection, extract path
5. Handle exit code 130 (ESC) → return ErrCancelled

### Display Format
```
main        → /Users/me/code/myapp
feature-auth → /Users/me/code/myapp-feature-auth
* feature-x   → /Users/me/code/myapp-feature-x  (current)
```

Current worktree marked with `*`, shown but not selectable (filter out after selection or use fzf preview).

### Verification
- Running Pick() opens fzf with worktree list
- Selection returns correct path
- ESC returns ErrCancelled

---

## Phase 4: Environment Symlinking

### internal/env/symlink.go

```go
package env

// Patterns to symlink from source to destination worktree
var DefaultPatterns = []string{
    ".env",
    ".env.*",
    ".envrc",
    ".claude",
}

// Symlink creates symlinks in dst for files matching patterns in src
// Skips files that don't exist in src
// Returns list of created symlinks
func Symlink(src, dst string, patterns []string) ([]string, error)
```

### Implementation

```go
func Symlink(src, dst string, patterns []string) ([]string, error) {
    var created []string

    for _, pattern := range patterns {
        matches, _ := filepath.Glob(filepath.Join(src, pattern))
        for _, srcPath := range matches {
            name := filepath.Base(srcPath)
            dstPath := filepath.Join(dst, name)

            // Skip if already exists
            if _, err := os.Lstat(dstPath); err == nil {
                continue
            }

            // Create relative symlink
            relPath, _ := filepath.Rel(dst, srcPath)
            if err := os.Symlink(relPath, dstPath); err != nil {
                return created, err
            }
            created = append(created, name)
        }
    }
    return created, nil
}
```

### Edge Cases
- `.claude` is a directory → symlink works the same
- `.env.local`, `.env.development` → matched by `.env.*`
- File already exists in dst → skip (don't overwrite)

### Verification
- Creating new worktree symlinks all matching files
- Symlinks are relative (portable)
- Missing files in source are skipped silently

---

## Phase 5: CLI Commands

### internal/cmd/root.go (updated)

```go
var rootCmd = &cobra.Command{
    Use:   "gw",
    Short: "Git worktree manager with fzf",
    RunE:  runSwitch, // Default action is interactive switch
}
```

### internal/cmd/switch.go

```go
func runSwitch(cmd *cobra.Command, args []string) error {
    worktrees, err := worktree.List()
    if err != nil {
        return err
    }

    selected, err := fzf.Pick(worktrees, "Switch to: ")
    if errors.Is(err, fzf.ErrCancelled) {
        os.Exit(2) // User cancelled
    }
    if err != nil {
        return err
    }

    // Output path for shell function to cd
    fmt.Println(selected)

    if claudeFlag {
        // Output special marker for shell function
        fmt.Println("__GW_LAUNCH_CLAUDE__")
    }

    return nil
}
```

### internal/cmd/new.go

```go
var newCmd = &cobra.Command{
    Use:   "new <branch>",
    Short: "Create new worktree",
    Args:  cobra.ExactArgs(1),
    RunE:  runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
    branch := args[0]

    current, err := worktree.Current()
    if err != nil {
        return err
    }

    newPath, err := worktree.SiblingPath(branch)
    if err != nil {
        return err
    }

    if err := worktree.Add(newPath, branch); err != nil {
        return err
    }

    // Symlink env files
    created, err := env.Symlink(current.Path, newPath, env.DefaultPatterns)
    if err != nil {
        // Log warning but don't fail
        fmt.Fprintf(os.Stderr, "Warning: failed to symlink some env files: %v\n", err)
    }
    if len(created) > 0 {
        fmt.Fprintf(os.Stderr, "Symlinked: %s\n", strings.Join(created, ", "))
    }

    // Output path for shell function
    fmt.Println(newPath)

    if claudeFlag {
        fmt.Println("__GW_LAUNCH_CLAUDE__")
    }

    return nil
}
```

### internal/cmd/ls.go

```go
var lsCmd = &cobra.Command{
    Use:   "ls",
    Short: "List worktrees",
    RunE:  runLs,
}

func runLs(cmd *cobra.Command, args []string) error {
    worktrees, err := worktree.List()
    if err != nil {
        return err
    }

    current, _ := worktree.Current()

    for _, wt := range worktrees {
        marker := "  "
        if current != nil && wt.Path == current.Path {
            marker = "* "
        }
        fmt.Printf("%s%s\t%s\n", marker, wt.Branch, wt.Path)
    }

    return nil
}
```

### internal/cmd/rm.go

```go
var rmCmd = &cobra.Command{
    Use:   "rm",
    Short: "Remove worktree",
    RunE:  runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
    worktrees, err := worktree.List()
    if err != nil {
        return err
    }

    // Filter out current worktree
    current, _ := worktree.Current()
    var removable []worktree.Worktree
    for _, wt := range worktrees {
        if current == nil || wt.Path != current.Path {
            removable = append(removable, wt)
        }
    }

    selected, err := fzf.Pick(removable, "Remove: ")
    if errors.Is(err, fzf.ErrCancelled) {
        os.Exit(2)
    }
    if err != nil {
        return err
    }

    return worktree.Remove(selected)
}
```

### Updated Shell Function

```bash
gw() {
    local output
    local exit_code

    output=$(command gw "$@")
    exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        [[ -n "$output" ]] && echo "$output"
        return $exit_code
    fi

    # Parse output: first line is path, second line may be __GW_LAUNCH_CLAUDE__
    local path launch_claude
    path=$(echo "$output" | head -1)
    launch_claude=$(echo "$output" | grep -q "__GW_LAUNCH_CLAUDE__" && echo "1")

    if [[ -d "$path" ]]; then
        cd "$path"
        [[ -n "$launch_claude" ]] && claude
    elif [[ -n "$path" ]]; then
        echo "$path"
    fi
}
```

### Verification
- `gw` opens fzf, selection outputs path
- `gw new feature` creates worktree, symlinks env, outputs path
- `gw ls` lists all worktrees with current marked
- `gw rm` opens fzf (without current), removes selection
- `-c` flag triggers claude launch marker

---

## Phase 6: Homebrew Distribution

### Repository Structure

Create separate repo `homebrew-tap` or add to this repo:

```
homebrew-gw/
└── Formula/
    └── gw.rb
```

### Formula/gw.rb

```ruby
class Gw < Formula
  desc "Git worktree manager with fzf integration"
  homepage "https://github.com/1unoe/gw"
  url "https://github.com/1unoe/gw/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "PLACEHOLDER"
  license "MIT"

  depends_on "go" => :build
  depends_on "fzf"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/gw"
  end

  def caveats
    <<~EOS
      To enable directory switching, add this to your shell config:

      # ~/.bashrc or ~/.zshrc
      gw() {
        local output exit_code
        output=$(command gw "$@")
        exit_code=$?
        [[ $exit_code -ne 0 ]] && { [[ -n "$output" ]] && echo "$output"; return $exit_code; }
        local path=$(echo "$output" | head -1)
        local launch_claude=$(echo "$output" | grep -q "__GW_LAUNCH_CLAUDE__" && echo "1")
        [[ -d "$path" ]] && { cd "$path"; [[ -n "$launch_claude" ]] && claude; } || [[ -n "$path" ]] && echo "$path"
      }
    EOS
  end

  test do
    system "#{bin}/gw", "--help"
  end
end
```

### Release Process

1. Tag release: `git tag v0.1.0 && git push --tags`
2. Create GitHub release with tarball
3. Update formula with SHA256 of tarball
4. Users install: `brew tap 1unoe/gw && brew install gw`

### Verification
- `brew install --build-from-source Formula/gw.rb` succeeds locally
- Post-install caveats display shell function
- `brew test gw` passes

---

## File Summary

| File | Purpose |
|------|---------|
| `go.mod` | Go module definition |
| `cmd/gw/main.go` | Entry point |
| `internal/cmd/root.go` | Cobra root command, -c flag |
| `internal/cmd/switch.go` | Default fzf switch behavior |
| `internal/cmd/new.go` | `gw new <branch>` |
| `internal/cmd/ls.go` | `gw ls` |
| `internal/cmd/rm.go` | `gw rm` |
| `internal/worktree/worktree.go` | Git worktree operations |
| `internal/fzf/fzf.go` | FZF picker |
| `internal/env/symlink.go` | Env file symlinking |
| `Makefile` | Build/install targets |
| `Formula/gw.rb` | Homebrew formula |

## Dependencies

```
github.com/spf13/cobra v1.8.0
```

No other external dependencies—git and fzf are exec'd.
