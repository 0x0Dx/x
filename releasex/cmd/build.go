package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0x0Dx/x/releasex/internal/archiver"
	"github.com/0x0Dx/x/releasex/internal/builder"
	"github.com/0x0Dx/x/releasex/internal/checksums"
	"github.com/0x0Dx/x/releasex/internal/config"
	"github.com/spf13/cobra"
)

var buildDir string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build binaries",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if buildDir == "" {
			buildDir = "dist"
		}

		if err := os.MkdirAll(buildDir, 0o755); err != nil {
			return fmt.Errorf("failed to create dist dir: %w", err)
		}

		results, err := builder.Build(cfg, buildDir)
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		for _, a := range cfg.Archives {
			files := findBuildFiles(results, a.Builds)
			for _, f := range a.Files {
				files = append(files, filepath.Join(buildDir, f))
			}
			output := filepath.Join(buildDir, cfg.Project+"-"+a.ID+"."+a.Format)
			if err := archiver.Create(files, a.Format, output); err != nil {
				return fmt.Errorf("archive failed: %w", err)
			}
		}

		for _, c := range cfg.Checksums {
			files := findArchiveFiles(results, c.IDs, cfg.Archives, cfg.Project)
			if len(files) == 0 {
				files = extractPaths(results)
			}
			output := filepath.Join(buildDir, cfg.Project+"-"+c.IDs[0]+"-checksums.txt")
			if err := checksums.Generate(files, output); err != nil {
				return fmt.Errorf("checksums failed: %w", err)
			}
		}

		fmt.Println("Build complete!")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildDir, "dir", "d", "", "Output directory")
}

func findBuildFiles(results []builder.Result, ids []string) []string {
	var files []string
	for _, r := range results {
		files = append(files, r.Path)
	}
	return files
}

func findArchiveFiles(_ []builder.Result, _ []string, archives []config.Archive, project string) []string {
	var files []string
	for _, a := range archives {
		files = append(files, filepath.Join(buildDir, project+"-"+a.ID+"."+a.Format))
	}
	return files
}

func extractPaths(results []builder.Result) []string {
	var paths []string
	for _, r := range results {
		paths = append(paths, r.Path)
	}
	return paths
}
