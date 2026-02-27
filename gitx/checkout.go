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

		var gitCmd *exec.Cmd
		if createNew {
			gitCmd = exec.Command("git", "checkout", "-b", branch)
		} else {
			gitCmd = exec.Command("git", "checkout", branch)
		}
		gitCmd.Stdin = nil
		gitCmd.Stdout = nil
		gitCmd.Stderr = nil

		if err := gitCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(Success.Render("✓ Switched to '" + branch + "'"))
	},
}

func init() {
	checkoutCmd.Flags().BoolP("branch", "b", false, "Create and switch to new branch")
	rootCmd.AddCommand(checkoutCmd)
}
