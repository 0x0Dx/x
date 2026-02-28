// Package main is the main entry point for the mochii CLI.
package main

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/mochii/cmd"
)

// main is the entry point for the mochii CLI.
func main() {
	if len(os.Args) < 2 {
		cmd.PrintUsage()
		os.Exit(1)
	}

	cli, err := cmd.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := cli.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: close cli failed: %v\n", err)
	}

	var exitCode int

	switch os.Args[1] {
	case "init":
		fmt.Println("initialized")

	case "verify":
		if err := cli.Verify(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "switch-config":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing config file\n")
			os.Exit(1)
		}
		if err := cli.SwitchConfig(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "config":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing config file\n")
			os.Exit(1)
		}
		if err := cli.Config(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "profile":
		if err := cli.ListProfiles(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "gc":
		if err := cli.CollectGarbage(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "pull":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing url arguments\n")
			os.Exit(1)
		}
		if err := cli.PullPrebuilts(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "push":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing directory argument\n")
			os.Exit(1)
		}
		if err := cli.PushPrebuilts(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "-h", "--help", "help":
		cmd.PrintUsage()

	default:
		fmt.Fprint(os.Stderr, "error: unknown command: ")
		fmt.Println(os.Args[1])
		cmd.PrintUsage()
		exitCode = 1
	}

	os.Exit(exitCode)
}
