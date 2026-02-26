package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Avatar v0.1.0\n")
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)

	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfg := GetConfig()
	token := GetToken()

	imageData, ext, err := GetImageData(cfg.Image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	provider, err := GetProvider(cfg.Provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := provider.ChangeAvatar(imageData, ext, token); err != nil {
		fmt.Fprintf(os.Stderr, "Error changing avatar: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully changed avatar on %s\n", cfg.Provider)
}
