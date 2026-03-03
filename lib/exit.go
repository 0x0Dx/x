// Package lib provides common utilities for CLI tools in this monorepo.
package lib

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Run executes the given cobra command and handles errors.
func Run(cmd *cobra.Command) error {
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}
	return nil
}

// MustGetEnv returns an environment variable or exits with an error.
func MustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is not set\n", key)
		os.Exit(1)
	}
	return val
}

// ExitOnError prints the error and exits with code 1 if err is not nil.
func ExitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
