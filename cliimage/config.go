package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

var rootCmd = &cobra.Command{
	Use:   "cliimage",
	Short: "A terminal image viewer",
	Long:  "Allows to render images in the Terminal",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.InputFile == "" {
			return fmt.Errorf("input file is required (use -i)")
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.InputFile, "input", "i", "", "input image file (required)")
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "output file (optional, defaults to stdout)")
	rootCmd.Flags().IntVarP(&cfg.Width, "width", "w", 0, "output width in characters")
	rootCmd.Flags().IntVar(&cfg.Height, "height", 0, "output height in characters")
	rootCmd.Flags().IntVarP(&cfg.Threshold, "threshold", "t", 128, "Threshold level for luminance (0-255)")
	rootCmd.Flags().StringVarP(&cfg.Pixelation, "pixelation", "p", "half", `Pixelation mode (default "half"):
  half    - Half blocks with 24-bit color
  quarter - Quarter blocks for higher resolution`)
	rootCmd.Flags().BoolVarP(&cfg.Dither, "dither", "d", false, "Apply Floyd-Steinberg dithering")
	rootCmd.Flags().BoolVarP(&cfg.NoBlockSymbols, "noblock", "b", false, "Use only half blocks (no block symbol matching)")
	rootCmd.Flags().BoolVarP(&cfg.Invert, "invert", "r", false, "Invert colors")
	rootCmd.Flags().IntVar(&cfg.Scale, "scale", 1, "Scale factor for rendering (1 = 1 character per 2x2 pixels)")

	rootCmd.MarkFlagRequired("input")
}

func Execute() error {
	return rootCmd.Execute()
}

func GetConfig() Config {
	return cfg
}
