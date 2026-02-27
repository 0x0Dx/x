package main

import (
	"fmt"
	"os"

	"github.com/0x0Dx/x/mochii/cmd"
)

// main is the entry point for the mochii CLI.
func main() {
	// Show usage if no command provided
	if len(os.Args) < 2 {
		cmd.PrintUsage()
		os.Exit(1)
	}

	// Initialize CLI
	cli, err := cmd.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	var exitCode int

	// Route commands
	switch os.Args[1] {
	case "init":
		fmt.Println("initialized")

	case "verify":
		if err := cli.Verify(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "getpkg":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing hash argument\n")
			os.Exit(1)
		}
		if err := cli.GetPkg(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "delpkg":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing hash argument\n")
			os.Exit(1)
		}
		if err := cli.Delete(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "listinst":
		if err := cli.ListInstalled(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing hash argument\n")
			os.Exit(1)
		}
		if err := cli.Run(os.Args[2], os.Args[3:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "regfile":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing file argument\n")
			os.Exit(1)
		}
		if err := cli.RegisterFile(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "regurl":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "error: missing hash or url argument\n")
			os.Exit(1)
		}
		if err := cli.RegisterURL(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "fetch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: missing url argument\n")
			os.Exit(1)
		}
		if err := cli.Fetch(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "profile":
		if err := cli.ListProfiles(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			exitCode = 1
		}

	case "switch":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "error: missing hash or path argument\n")
			os.Exit(1)
		}
		if err := cli.SwitchProfile(os.Args[2], os.Args[3]); err != nil {
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
		fmt.Fprintf(os.Stderr, "error: unknown command: %s\n", os.Args[1])
		cmd.PrintUsage()
		exitCode = 1
	}

	os.Exit(exitCode)
}
