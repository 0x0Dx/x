// Package cmd provides releasex CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd is the root cobra command.
var RootCmd = &cobra.Command{
	Use:   "releasex",
	Short: "Simple release tool for Go projects",
	Long:  "Build, archive, and release Go projects to GitHub.",
}

// Execute runs the root command.
func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}
	return nil
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "releasex.yaml", "Config file")
}
