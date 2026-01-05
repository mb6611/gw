package cmd

import (
    "errors"
    "fmt"
    "os"

    "github.com/1unoe/gw/internal/fzf"
    "github.com/1unoe/gw/internal/worktree"
    "github.com/spf13/cobra"
)

func runSwitch(cmd *cobra.Command, args []string) error {
    worktrees, err := worktree.List()
    if err != nil {
        return err
    }

    current, _ := worktree.Current()
    currentPath := ""
    if current != nil {
        currentPath = current.Path
    }

    selected, err := fzf.Pick(worktrees, currentPath, "Switch to: ")
    if errors.Is(err, fzf.ErrCancelled) {
        os.Exit(2)
    }
    if errors.Is(err, fzf.ErrNotFound) {
        return fmt.Errorf("fzf is required but not found in PATH")
    }
    if err != nil {
        return err
    }

    // Output path for shell function to cd
    fmt.Println(selected)

    if ClaudeFlag {
        fmt.Println("__GW_LAUNCH_CLAUDE__")
    }

    return nil
}
