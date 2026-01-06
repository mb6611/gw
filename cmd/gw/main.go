package main

import "github.com/mb6611/gw/internal/cmd"

var version = "dev"

func main() {
    cmd.SetVersion(version)
    cmd.Execute()
}
