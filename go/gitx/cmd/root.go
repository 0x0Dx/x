// Package cmd provides gitx commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var verbose bool

// RootCmd is the root cobra command.
var RootCmd = &cobra.Command{
	Use:   "gitx",
	Short: "A simple, opinionated git wrapper",
	Long:  "Gitx is a simple, opinionated git wrapper with shorter commands and nicer output.",
}

// Execute runs the root command.
func Execute() error {
	err := RootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
