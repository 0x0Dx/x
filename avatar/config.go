// Package main provides a CLI tool to change avatars for various services.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Config struct {
	Provider string
	Image    string
	Token    string
}

var cfg Config

var rootCmd = &cobra.Command{
	Use:   "avatar",
	Short: "Change avatar for various services",
	Long:  "A CLI tool to change avatars for GitHub, Discord, Steam, and more",
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfg.Image == "" {
			return fmt.Errorf("image path or URL is required (use -i)")
		}
		if cfg.Provider == "" {
			return fmt.Errorf("provider is required (use -p)")
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.Provider, "provider", "p", "", "provider: github, discord, steam")
	rootCmd.Flags().StringVarP(&cfg.Image, "image", "i", "", "path to image file or URL")
	rootCmd.Flags().StringVarP(&cfg.Token, "token", "t", "", "API token (or set via environment variable)")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable debug output")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
}

func GetConfig() Config {
	return cfg
}

func GetToken() string {
	if cfg.Token != "" {
		return cfg.Token
	}
	return os.Getenv("AVATAR_TOKEN")
}
