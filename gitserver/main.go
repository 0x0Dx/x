// Package main is the entry point for gitserver.
package main

import (
	"os"

	"github.com/0x0Dx/x/gitserver/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
