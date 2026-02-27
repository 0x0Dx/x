package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "ci <message>",
	Short: "Stage all and commit",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		msg := strings.Join(args, " ")

		addCmd := exec.Command("git", "add", "-A")
		if err := addCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error staging: "+err.Error()))
			os.Exit(1)
		}

		commitCmd := exec.Command("git", "commit", "-m", msg)
		if err := commitCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error committing: "+err.Error()))
			os.Exit(1)
		}
		fmt.Println(Success.Render("✓ Committed: " + msg))
	},
}

var statusCmd = &cobra.Command{
	Use:   "st",
	Short: "Show working tree status",
	Run: func(_ *cobra.Command, _ []string) {
		cmd := exec.Command("git", "status", "--short")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error: "+err.Error()))
			os.Exit(1)
		}
	},
}

var branchCmd = &cobra.Command{
	Use:   "br",
	Short: "List branches",
	Run: func(_ *cobra.Command, _ []string) {
		cmd := exec.Command("git", "branch")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, Error.Render("Error: "+err.Error()))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(branchCmd)
}
