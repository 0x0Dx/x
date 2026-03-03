package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/0x0Dx/x/cliimage/internal/blocks"
	"github.com/spf13/cobra"
)

// RootCmd is the root command for the cliimage CLI.
var RootCmd = &cobra.Command{
	Use:   "cliimage",
	Short: "Terminal image viewer",
}

// Execute runs the root command.
func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}
	return nil
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
