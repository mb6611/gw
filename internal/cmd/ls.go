package cmd

import (
    "fmt"

    "github.com/mb6611/gw/internal/worktree"
    "github.com/spf13/cobra"
)

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
        if wt.Bare {
            continue
        }

        marker := "  "
        if current != nil && wt.Path == current.Path {
            marker = "* "
        }
        fmt.Printf("%s%s\t%s\n", marker, wt.Branch, wt.Path)
    }

    return nil
}
