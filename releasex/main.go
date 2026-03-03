// releasex is a simple release tool for Go projects.
package main

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/releasex/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
