// Package main provides gitx commands.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "co <branch> [-b]",
	Short: "Checkout a branch",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cc *cobra.Command, args []string) {
		branch := args[0]
		createNew, _ := cc.Flags().GetBool("branch")

		var cmd *exec.Cmd
		if createNew {
			cmd = exec.Command("git", "checkout", "-b", branch)
		} else {
			cmd = exec.Command("git", "checkout", branch)
		}
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Switched to '" + branch + "'")
	},
}

func init() {
	checkoutCmd.Flags().BoolP("branch", "b", false, "Create and switch to new branch")
	rootCmd.AddCommand(checkoutCmd)
}
