// Package cmd provides gitx commands.
package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [-c] [--] [files...]",
	Short: "Show changes",
	Args:  cobra.ArbitraryArgs,
	Run: func(cc *cobra.Command, args []string) {
		staged, _ := cc.Flags().GetBool("cached")

		gitArgs := []string{"diff"}
		if staged {
			gitArgs = append(gitArgs, "--cached")
		}
		if len(args) > 0 {
			gitArgs = append(gitArgs, "--")
			gitArgs = append(gitArgs, args...)
		}

		cmd := exec.Command("git", gitArgs...)
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
