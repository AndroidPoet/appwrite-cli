package main

import (
	"fmt"
	"os"

	"github.com/AndroidPoet/appwrite-cli/cmd/appwrite-cli/commands"
)

// Version information (set by build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	commands.SetVersionInfo(version, commit, date)
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
