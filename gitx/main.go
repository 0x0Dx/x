// Package main is the entry point for gitx.
package main

import (
	"os"

	"github.com/0x0Dx/x/gitx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
