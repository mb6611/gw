package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path   string // Absolute path to worktree
	Branch string // Branch name (or HEAD commit if detached)
	Bare   bool   // Is this the bare repo
}

// List returns all worktrees for the repo containing the current directory
func List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []Worktree
	var current *Worktree

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "branch ") {
			if current != nil {
				branch := strings.TrimPrefix(line, "branch ")
				current.Branch = strings.TrimPrefix(branch, "refs/heads/")
			}
		} else if line == "bare" {
			if current != nil {
				current.Bare = true
			}
		} else if strings.HasPrefix(line, "HEAD ") {
			// Detached HEAD - use commit hash
			if current != nil && current.Branch == "" {
				current.Branch = strings.TrimPrefix(line, "HEAD ")
			}
		}
	}

	// Add the last worktree if exists
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing worktree list: %w", err)
	}

	return worktrees, nil
}

// Add creates a new worktree at the given path for the given branch
// If branch doesn't exist, creates it from current HEAD
func Add(path, branch string) error {
	// Check if branch exists
	checkCmd := exec.Command("git", "rev-parse", "--verify", branch)
	err := checkCmd.Run()

	var addCmd *exec.Cmd
	if err == nil {
		// Branch exists
		addCmd = exec.Command("git", "worktree", "add", path, branch)
	} else {
		// Branch doesn't exist, create it
		addCmd = exec.Command("git", "worktree", "add", "-b", branch, path)
	}

	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add worktree: %w\n%s", err, output)
	}

	return nil
}

// Remove deletes a worktree (must not be current directory)
func Remove(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w\n%s", err, output)
	}
	return nil
}

// Current returns the worktree for the current directory
func Current() (*Worktree, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	worktrees, err := List()
	if err != nil {
		return nil, err
	}

	// Find the worktree that contains the current directory
	for _, wt := range worktrees {
		// Check if cwd is within this worktree
		rel, err := filepath.Rel(wt.Path, cwd)
		if err != nil {
			continue
		}
		// If the relative path doesn't start with "..", we're inside this worktree
		if !strings.HasPrefix(rel, "..") {
			return &wt, nil
		}
	}

	return nil, fmt.Errorf("current directory is not in a git worktree")
}

// RepoName returns the base repository name (used for sibling naming)
func RepoName() (string, error) {
	// Get the git directory
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git directory: %w", err)
	}

	gitDir := strings.TrimSpace(string(output))

	// If it's just ".git", we're in the main worktree
	if gitDir == ".git" {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get toplevel: %w", err)
		}
		toplevel := strings.TrimSpace(string(output))
		return filepath.Base(toplevel), nil
	}

	// If it contains "/worktrees/", extract from the bare repo path
	if !strings.Contains(gitDir, "/worktrees/") {
		return "", fmt.Errorf("unexpected git directory format: %s", gitDir)
	}

	// gitDir looks like: /path/to/repo/.git/worktrees/branch-name
	// We want the base repo directory
	parts := strings.Split(gitDir, "/worktrees/")
	if len(parts) == 0 {
		return "", fmt.Errorf("failed to parse worktree path from: %s", gitDir)
	}

	bareGitDir := parts[0]
	repoDir := filepath.Dir(bareGitDir)
	name := filepath.Base(repoDir)

	// Strip any existing branch suffix (pattern: name-branchname)
	// Look for the last hyphen and check if what follows looks like a branch
	if idx := strings.LastIndex(name, "-"); idx != -1 {
		// Return the part before the last hyphen as the base name
		return name[:idx], nil
	}

	return name, nil
}

// SiblingPath generates the path for a new worktree as a sibling directory
func SiblingPath(branch string) (string, error) {
	repoName, err := RepoName()
	if err != nil {
		return "", err
	}

	current, err := Current()
	if err != nil {
		return "", err
	}

	parent := filepath.Dir(current.Path)
	return filepath.Join(parent, repoName+"-"+branch), nil
}
