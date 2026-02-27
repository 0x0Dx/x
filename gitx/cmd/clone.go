// Package cmd provides gitx commands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <repo> [dir]",
	Short: "Clone a repository",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		repo := args[0]
		var dir string
		if len(args) > 1 {
			dir = args[1]
		} else {
			parts := strings.Split(repo, "/")
			dir = strings.TrimSuffix(parts[len(parts)-1], ".git")
		}

		cmd := exec.Command("git", "clone", repo, dir)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		fmt.Println("Cloning into '" + dir + "'...")
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("✓ Cloned successfully")
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)
}
