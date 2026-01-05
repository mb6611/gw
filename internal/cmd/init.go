package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <shell>",
	Short: "Generate shell integration",
	Long:  "Generate shell integration script. Add 'eval \"$(gw init zsh)\"' to your .zshrc",
	Args:  cobra.ExactArgs(1),
	RunE:  runInit,
}

const zshInit = `gw() {
  local output exit_code
  output=$(command gw "$@")
  exit_code=$?

  if [[ $exit_code -ne 0 ]]; then
    [[ -n "$output" ]] && echo "$output"
    return $exit_code
  fi

  local path=$(echo "$output" | head -1)

  if [[ -d "$path" ]]; then
    builtin cd "$path"
    [[ "$output" == *"__GW_LAUNCH_CLAUDE__"* ]] && claude
  elif [[ -n "$path" ]]; then
    echo "$path"
  fi
}`

const bashInit = zshInit // Same syntax works for bash

const fishInit = `function gw
  set -l output (command gw $argv)
  set -l exit_code $status

  if test $exit_code -ne 0
    test -n "$output" && echo "$output"
    return $exit_code
  end

  set -l path (echo "$output" | head -1)

  if test -d "$path"
    cd "$path"
    string match -q "*__GW_LAUNCH_CLAUDE__*" "$output" && claude
  else if test -n "$path"
    echo "$path"
  end
end`

func runInit(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "zsh":
		fmt.Println(zshInit)
	case "bash":
		fmt.Println(bashInit)
	case "fish":
		fmt.Println(fishInit)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: zsh, bash, fish)", shell)
	}

	return nil
}
