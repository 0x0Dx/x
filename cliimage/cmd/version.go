package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("A terminal image viewer v0.1.0\n")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
