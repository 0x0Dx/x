package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "br [-d <branch>] [-c <branch>] [-s <branch>]",
	Short: "List, create, or delete branches",
	Run: func(cc *cobra.Command, _ []string) {
		deleteBranch, _ := cc.Flags().GetString("delete")
		createBranch, _ := cc.Flags().GetString("create")
		switchBranch, _ := cc.Flags().GetString("switch")

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

		if createBranch != "" {
			cmd := exec.Command("git", "branch", createBranch)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Println("✓ Created branch:", createBranch)
			return
		}

		if switchBranch != "" {
			cmd := exec.Command("git", "checkout", "-b", switchBranch)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Println("✓ Created and switched to branch:", switchBranch)
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

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Swap to previous branch",
	Run: func(_ *cobra.Command, _ []string) {
		cmd := exec.Command("git", "checkout", "-")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Swapped to previous branch")
	},
}

func init() {
	branchCmd.Flags().StringP("delete", "d", "", "Delete a branch")
	branchCmd.Flags().StringP("create", "c", "", "Create a branch")
	branchCmd.Flags().StringP("switch", "s", "", "Create and switch to a new branch")
	RootCmd.AddCommand(branchCmd)
	RootCmd.AddCommand(swapCmd)
}
