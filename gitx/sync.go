package main

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
			fmt.Fprintln(os.Stderr, Error.Render("Error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(Success.Render("✓ Pushed"))
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
			fmt.Fprintln(os.Stderr, Error.Render("Error: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(Success.Render("✓ Pulled"))
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull, rebase and push",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(Info.Render("Pulling..."))
		cmd := exec.Command("git", "pull", "--rebase")
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error pulling: "+err.Error()))
			os.Exit(1)
		}

		fmt.Println(Info.Render("Pushing..."))
		cmd = exec.Command("git", "push")
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error pushing: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(Success.Render("✓ Synced"))
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(syncCmd)
}
