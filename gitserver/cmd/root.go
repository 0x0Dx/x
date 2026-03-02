// Package cmd provides gitserver commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RootCmd is the root cobra command.
var RootCmd = &cobra.Command{
	Use:   "gitserver",
	Short: "A simple gitserver",
	Long:  "A simple gitserver",
}

// Execute runs the root command.
func Execute() error {
	err := RootCmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
}
