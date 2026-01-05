package fzf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mb6611/gw/internal/worktree"
)

// ErrCancelled is returned when the user cancels fzf selection (ESC)
var ErrCancelled = errors.New("selection cancelled")

// ErrNotFound is returned when fzf is not installed
var ErrNotFound = errors.New("fzf not found in PATH")

// Pick shows worktrees in fzf and returns the selected path
// Returns ("", ErrCancelled) if user presses ESC
// Returns ("", ErrNotFound) if fzf is not installed
func Pick(worktrees []worktree.Worktree, currentPath string, prompt string) (string, error) {
	// Check fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return "", ErrNotFound
	}

	// Format worktrees for display
	var lines []string

	for _, wt := range worktrees {
		if wt.Bare {
			continue // Skip bare repos
		}

		marker := "  "
		suffix := ""
		if wt.Path == currentPath {
			marker = "* "
			suffix = " (current)"
		}

		// Format: "* branch-name  → /path/to/worktree (current)"
		display := fmt.Sprintf("%s%-20s → %s%s", marker, wt.Branch, wt.Path, suffix)
		lines = append(lines, display)
	}

	if len(lines) == 0 {
		return "", errors.New("no worktrees available")
	}

	// Run fzf
	cmd := exec.Command("fzf",
		"--prompt", prompt,
		"--height", "40%",
		"--reverse",
		"--ansi",
	)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		// Exit code 130 means user pressed ESC
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				return "", ErrCancelled
			}
		}
		return "", fmt.Errorf("fzf error: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", ErrCancelled
	}

	// Extract path from selection by parsing after "→"
	parts := strings.SplitN(selected, "→", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected selection format: %s", selected)
	}

	path := strings.TrimSpace(parts[1])
	// Remove " (current)" suffix if present
	path = strings.TrimSuffix(path, " (current)")

	return path, nil
}
