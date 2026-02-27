// Package cmd provides gitx commands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "st",
	Short: "Show working tree status",
	Run: func(_ *cobra.Command, _ []string) {
		if verbose {
			fmt.Println("Running: git status --short")
		}
		cmd := exec.Command("git", "status", "--short")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
