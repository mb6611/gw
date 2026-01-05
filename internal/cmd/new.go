package cmd

import (
    "fmt"
    "os"
    "strings"

    "github.com/mb6611/gw/internal/env"
    "github.com/mb6611/gw/internal/worktree"
    "github.com/spf13/cobra"
)

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
        fmt.Fprintf(os.Stderr, "Warning: failed to symlink some env files: %v\n", err)
    }
    if len(created) > 0 {
        fmt.Fprintf(os.Stderr, "Symlinked: %s\n", strings.Join(created, ", "))
    }

    // Output path for shell function
    fmt.Println(newPath)

    if ClaudeDangerousFlag {
        fmt.Println("__GW_LAUNCH_CLAUDE_DANGEROUS__")
    } else if ClaudeFlag {
        fmt.Println("__GW_LAUNCH_CLAUDE__")
    }

    return nil
}
