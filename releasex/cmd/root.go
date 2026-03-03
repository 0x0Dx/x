// Package cmd provides releasex CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "Config file (default: ReleaseX.yaml)")
	RootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if cfgFile != "" {
			return nil
		}
		cfgFiles := []string{"ReleaseX.yaml", "releasex.yaml"}
		for _, dir := range []string{".", ".."} {
			for _, name := range cfgFiles {
				path := filepath.Join(dir, name)
				if _, err := os.Stat(path); err == nil {
					cfgFile = path
					return nil
				}
			}
		}
		return fmt.Errorf("config file not found (tried ReleaseX.yaml, releasex.yaml)")
	}
}

// GetConfigPath returns the resolved config file path.
func GetConfigPath() string {
	abs, _ := filepath.Abs(cfgFile)
	return abs
}
