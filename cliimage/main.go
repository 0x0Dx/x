// Package main is the entry point for cliimage.
package main

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0x0Dx/x/cliimage/internal/blocks"
	"github.com/0x0Dx/x/cliimage/internal/config"
	"github.com/0x0Dx/x/cliimage/internal/renderer"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("A terminal image viewer v0.1.0\n")
	},
}

func main() {
	cobra.OnInitialize()

	config.RootCmd.AddCommand(versionCmd)

	if err := config.Execute(); err != nil {
		os.Exit(1)
	}

	cfg := config.GetConfig()

	file := openInputFile(cfg.InputFile)
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	img, format, err := image.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
		os.Exit(1)
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
			os.Exit(1)
		}
		fmt.Printf("Successfully converted %s (%s format) to %s\n", cfg.InputFile, format, outputPath)
	} else {
		fmt.Print(result)
	}
}

func getTerminalWidth() int {
	ctx, cancel := context.WithTimeout(context.Background(), 10000000000)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tput", "cols")
	out, _ := cmd.Output()
	if w, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil && w > 0 {
		if w < 120 {
			w *= 2
		}
		return w
	}
	return 80
}

func openInputFile(path string) *os.File {
	if path == "-" {
		return os.Stdin
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	return file
}

func parseSymbol(platform string) blocks.Symbol {
	switch platform {
	case "quarter":
		return blocks.SymbolQuarter
	case "all":
		return blocks.SymbolAll
	default:
		return blocks.SymbolHalf
	}
}
