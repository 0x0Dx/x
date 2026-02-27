// Package config provides CLI configuration and cobra commands.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Config holds the CLI configuration.
type Config struct {
	InputFile      string
	OutputFile     string
	Width          int
	Height         int
	Threshold      int
	Pixelation     string
	Dither         bool
	NoBlockSymbols bool
	Invert         bool
	Scale          int
}

var cfg Config

// RootCmd is the root cobra command.
var RootCmd = &cobra.Command{
	Use:   "cliimage",
	Short: "A terminal image viewer",
	Long:  "Allows to render images in the Terminal",
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfg.InputFile == "" {
			return fmt.Errorf("input file is required (use -i)")
		}
		return nil
	},
}

func init() {
	RootCmd.Flags().StringVarP(&cfg.InputFile, "input", "i", "", "input image file (required)")
	RootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "output file (optional, defaults to stdout)")
	RootCmd.Flags().IntVarP(&cfg.Width, "width", "w", 0, "output width in characters")
	RootCmd.Flags().IntVar(&cfg.Height, "height", 0, "output height in characters")
	RootCmd.Flags().IntVarP(&cfg.Threshold, "threshold", "t", 128, "Threshold level for luminance (0-255)")
	RootCmd.Flags().StringVarP(&cfg.Pixelation, "pixelation", "p", "half", `Pixelation mode (default "half"):
  half    - Half blocks with 24-bit color
  quarter - Quarter blocks for higher resolution`)
	RootCmd.Flags().BoolVarP(&cfg.Dither, "dither", "d", false, "Apply Floyd-Steinberg dithering")
	RootCmd.Flags().BoolVarP(&cfg.NoBlockSymbols, "noblock", "b", false, "Use only half blocks (no block symbol matching)")
	RootCmd.Flags().BoolVarP(&cfg.Invert, "invert", "r", false, "Invert colors")
	RootCmd.Flags().IntVar(&cfg.Scale, "scale", 1, "Scale factor for rendering (1 = 1 character per 2x2 pixels)")

	if err := RootCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

// Execute runs the root cobra command.
func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
}

// GetConfig returns the current configuration.
func GetConfig() Config {
	return cfg
}
