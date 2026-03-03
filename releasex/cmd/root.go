package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "releasex",
	Short: "Simple release tool for Go projects",
	Long:  "Build, archive, and release Go projects to GitHub.",
}

func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}
	return nil
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "releasex.yaml", "Config file")
}
