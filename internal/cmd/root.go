package cmd

import (
    "os"
    "github.com/spf13/cobra"
)

// ClaudeFlag indicates whether to launch claude after switching
var ClaudeFlag bool

var rootCmd = &cobra.Command{
    Use:   "gw",
    Short: "Git worktree manager with fzf",
    RunE:  runSwitch,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().BoolVarP(&ClaudeFlag, "claude", "c", false, "Launch claude after switching")

    rootCmd.AddCommand(newCmd)
    rootCmd.AddCommand(lsCmd)
    rootCmd.AddCommand(rmCmd)
    rootCmd.AddCommand(initCmd)
}
