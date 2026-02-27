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
			fmt.Fprintln(os.Stderr, "Error staging:", err)
			os.Exit(1)
		}

		cm := exec.Command("git", "commit", "-m", msg)
		if err := cm.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error committing:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Committed:", msg)
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
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var branchCmd = &cobra.Command{
	Use:   "br [-d <branch>]",
	Short: "List or delete branches",
	Run: func(cc *cobra.Command, _ []string) {
		deleteBranch, _ := cc.Flags().GetString("delete")

		if deleteBranch != "" {
			cmd := exec.Command("git", "branch", "-D", deleteBranch)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Println("✓ Deleted branch:", deleteBranch)
			return
		}

		cmd := exec.Command("git", "branch")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	branchCmd.Flags().StringP("delete", "d", "", "Delete a branch")
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(branchCmd)
}
