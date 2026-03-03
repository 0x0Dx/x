package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate config file",
	Run: func(_ *cobra.Command, _ []string) {
		content := `# releasex.yaml - Simple release config
project: myapp
version: v0.1.0

builds:
  - id: main
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    main: ./cmd/app
    binary: myapp
    ldflags: "-s -w"

archives:
  - id: default
    builds: [main]
    format: tar.gz

checksums:
  - ids: [default]

# github:
#   owner: myorg
#   repo: myrepo
#   version: tag
#   draft: false
`

		if err := os.WriteFile("releasex.yaml", []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Println("Created releasex.yaml")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
