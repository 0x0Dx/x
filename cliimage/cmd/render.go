// Package cmd provides CLI commands for cliimage.
package cmd

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"  // Register GIF format decoders.
	_ "image/jpeg" // Register JPEG format decoders.
	_ "image/png"  // Register PNG format decoders.

	"github.com/0x0Dx/x/cliimage/internal/renderer"
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render an image to the terminal",
	RunE:  runRender,
}

func init() {
	RootCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringVarP(&cfg.InputFile, "input", "i", "", "input image file (required)")
	renderCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "output file (optional, defaults to stdout)")
	renderCmd.Flags().IntVarP(&cfg.Width, "width", "w", 0, "output width in characters")
	renderCmd.Flags().IntVar(&cfg.Height, "height", 0, "output height in characters")
	renderCmd.Flags().IntVarP(&cfg.Threshold, "threshold", "t", 128, "Threshold level for luminance (0-255)")
	renderCmd.Flags().StringVarP(&cfg.Pixelation, "pixelation", "p", "half", `Pixelation mode (default "half"):
  half    - Half blocks with 24-bit color
  quarter - Quarter blocks for higher resolution`)
	renderCmd.Flags().BoolVarP(&cfg.Dither, "dither", "d", false, "Apply Floyd-Steinberg dithering")
	renderCmd.Flags().BoolVarP(&cfg.NoBlockSymbols, "noblock", "b", false, "Use only half blocks (no block symbol matching)")
	renderCmd.Flags().BoolVarP(&cfg.Invert, "invert", "r", false, "Invert colors")
	renderCmd.Flags().IntVar(&cfg.Scale, "scale", 1, "Scale factor for rendering (1 = 1 character per 2x2 pixels)")

	if err := renderCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func runRender(_ *cobra.Command, _ []string) error {
	file := openInputFile(cfg.InputFile)
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	img, format, err := image.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
		return fmt.Errorf("failed to decode image: %w", err)
	}

	if cfg.Width == 0 {
		cfg.Width = getTerminalWidth()
	}

	pixelation := strings.ToLower(cfg.Pixelation)

	symbol := parseSymbol(pixelation)

	r := renderer.New().
		Width(cfg.Width).
		Height(cfg.Height).
		Threshold(cfg.Threshold).
		Symbol(symbol).
		Dither(cfg.Dither).
		IgnoreBlockSymbols(cfg.NoBlockSymbols).
		InvertColors(cfg.Invert).
		Scale(cfg.Scale)

	result := r.Render(img)

	if cfg.OutputFile != "" {
		outputPath := filepath.Clean(cfg.OutputFile)
		//nolint:gosec // G703: false positive - path is cleaned by filepath.Clean
		if err := os.WriteFile(outputPath, []byte(result), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Successfully converted %s (%s format) to %s\n", cfg.InputFile, format, outputPath)
	} else {
		fmt.Print(result)
	}

	return nil
}

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
