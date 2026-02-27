// Package cmd provides gitx commands.
package cmd

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
		addCmd.Stdin = nil
		addCmd.Stdout = nil
		addCmd.Stderr = nil
		_ = addCmd.Run()

		cm := exec.Command("git", "commit", "-m", msg)
		cm.Stdin = nil
		cm.Stdout = nil
		cm.Stderr = nil
		if err := cm.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Committed:", msg)
	},
}

func init() {
	RootCmd.AddCommand(commitCmd)
}
