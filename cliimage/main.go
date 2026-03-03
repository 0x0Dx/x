// Package main is the entry point for cliimage.
package main

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/cliimage/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
