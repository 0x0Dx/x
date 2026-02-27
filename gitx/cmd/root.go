package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the root cobra command.
var RootCmd = &cobra.Command{
	Use:   "gitx",
	Short: "A simple, opinionated git wrapper",
}

// Execute runs the root command.
func Execute() error {
	return RootCmd.Execute()
}
