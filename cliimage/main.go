package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("A terminal image viewer v0.1.0\n")
	},
}

func main() {
	cobra.OnInitialize()

	rootCmd.AddCommand(versionCmd)

	if err := Execute(); err != nil {
		os.Exit(1)
	}

	cfg := GetConfig()

	file := openInputFile(cfg.InputFile)
	defer file.Close()

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

	renderer := New().
		Width(cfg.Width).
		Height(cfg.Height).
		Threshold(cfg.Threshold).
		Symbol(symbol).
		Dither(cfg.Dither).
		IgnoreBlockSymbols(cfg.NoBlockSymbols).
		InvertColors(cfg.Invert).
		Scale(cfg.Scale)

	result := renderer.Render(img)

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, []byte(result), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully converted %s (%s format) to %s\n", cfg.InputFile, format, cfg.OutputFile)
	} else {
		fmt.Print(result)
	}
}

func getTerminalWidth() int {
	cmd := exec.Command("tput", "cols")
	out, _ := cmd.Output()
	if w, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil && w > 0 {
		// Kitty/modern terminals report half width
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

func parseSymbol(platform string) Symbol {
	switch platform {
	case "quarter":
		return SymbolQuarter
	case "all":
		return SymbolAll
	default:
		return SymbolHalf
	}
}
