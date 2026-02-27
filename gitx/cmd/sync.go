// Package cmd provides gitx commands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push to remote",
	Run: func(_ *cobra.Command, _ []string) {
		cmd := exec.Command("git", "push")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Pushed")
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull from remote",
	Run: func(_ *cobra.Command, _ []string) {
		cmd := exec.Command("git", "pull", "--rebase")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Pulled")
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull, rebase and push",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Pulling...")
		cmd := exec.Command("git", "pull", "--rebase")
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error pulling:", err)
			os.Exit(1)
		}

		fmt.Println("Pushing...")
		cmd = exec.Command("git", "push")
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error pushing:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Synced")
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(syncCmd)
}
