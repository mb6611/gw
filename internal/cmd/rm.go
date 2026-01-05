package cmd

import (
    "errors"
    "fmt"
    "os"

    "github.com/1unoe/gw/internal/fzf"
    "github.com/1unoe/gw/internal/worktree"
    "github.com/spf13/cobra"
)

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

    // Filter out current worktree and bare repos
    current, _ := worktree.Current()
    var removable []worktree.Worktree
    for _, wt := range worktrees {
        if wt.Bare {
            continue
        }
        if current == nil || wt.Path != current.Path {
            removable = append(removable, wt)
        }
    }

    if len(removable) == 0 {
        return fmt.Errorf("no worktrees available to remove")
    }

    selected, err := fzf.Pick(removable, "", "Remove: ")
    if errors.Is(err, fzf.ErrCancelled) {
        os.Exit(2)
    }
    if errors.Is(err, fzf.ErrNotFound) {
        return fmt.Errorf("fzf is required but not found in PATH")
    }
    if err != nil {
        return err
    }

    return worktree.Remove(selected)
}
