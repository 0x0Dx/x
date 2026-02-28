// Package main is the entry point for goreviewer.
package main

import (
	"os"

	"github.com/0x0Dx/x/goreviewer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
