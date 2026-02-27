package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo [--hard] [commit]",
	Short: "Undo changes (amend last commit or reset to commit)",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cc *cobra.Command, args []string) {
		hard, _ := cc.Flags().GetBool("hard")

		if len(args) == 0 {
			cmd := exec.Command("git", "commit", "--amend", "--no-edit")
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Println("✓ Amended last commit")
			return
		}

		mode := "soft"
		if hard {
			mode = "hard"
		}

		cmd := exec.Command("git", "reset", "--"+mode, args[0])
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Reset to", args[0], "["+mode+"]")
	},
}

func init() {
	undoCmd.Flags().BoolP("hard", "", false, "Hard reset (discards changes)")
	RootCmd.AddCommand(undoCmd)
}
