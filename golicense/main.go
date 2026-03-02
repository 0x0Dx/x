// Package main provides a CLI tool for generating license files.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/0x0Dx/x/golicense/licenses"
)

var (
	name    = flag.String("name", "", "Name of the person licensing the software")
	email   = flag.String("email", "", "Email of the person licensing the software")
	out     = flag.Bool("out", false, "Write output to a file instead of stdout")
	showAll = flag.Bool("show", false, "Show all available licenses")
)

func init() {
	flag.Usage = func() {
		//nolint:gosec
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <license>\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nLicense name can be one of: ")
		names, _ := licenses.List()
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(names, ", "))
		fmt.Fprintf(os.Stderr, "\nIf name/email not provided, they are read from git config.\n")
		os.Exit(2)
	}
}

// getGitConfig retrieves a value from git config.
func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", key)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getOutputFilename returns the appropriate filename for the given license kind.
func getOutputFilename(kind string) string {
	filenames := map[string]string{
		"unlicense": "UNLICENSE",
		"sqlite":    "BLESSING",
	}
	if name, ok := filenames[kind]; ok {
		return name
	}
	return "LICENSE"
}

func main() {
	flag.Parse()

	if *showAll {
		names, err := licenses.List()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Available licenses:")
		for _, n := range names {
			fmt.Printf("  %s\n", n)
		}
		return
	}

	if len(flag.Args()) != 1 {
		flag.Usage()
	}

	kind := flag.Arg(0)

	if !licenses.Has(kind) {
		fmt.Fprintf(os.Stderr, "invalid license: %s\n", kind)
		os.Exit(1)
	}

	author := *name
	if author == "" {
		author = getGitConfig("user.name")
	}

	authorEmail := *email
	if authorEmail == "" {
		authorEmail = getGitConfig("user.email")
	}

	var wr io.Writer
	if *out {
		filename := getOutputFilename(kind)
		fout, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := fout.Close(); err != nil {
				log.Printf("failed to close output file: %v", err)
			}
		}()
		wr = fout
	} else {
		wr = os.Stdout
		defer fmt.Println()
	}

	if err := licenses.Hydrate(kind, author, authorEmail, wr); err != nil {
		log.Fatal(err)
	}
}
