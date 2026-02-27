// Package cmd provides gitx commands.
package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [-c]",
	Short: "Show changes",
	Run: func(cc *cobra.Command, _ []string) {
		staged, _ := cc.Flags().GetBool("cached")

		args := []string{"diff"}
		if staged {
			args = append(args, "--cached")
		}

		cmd := exec.Command("git", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	diffCmd.Flags().BoolP("cached", "c", false, "Show staged changes")
	RootCmd.AddCommand(diffCmd)
}
