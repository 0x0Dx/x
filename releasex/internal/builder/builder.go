package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/0x0Dx/x/releasex/internal/config"
)

type Result struct {
	Path     string
	GoOS     string
	GoArch   string
	Checksum string
}

func Build(cfg *config.Config, dir, projectRoot string) ([]Result, error) {
	var results []Result

	for _, b := range cfg.Builds {
		for _, goos := range b.GoOS {
			for _, goarch := range b.GoArch {
				r, err := buildOne(b, goos, goarch, dir, projectRoot)
				if err != nil {
					return nil, err
				}
				results = append(results, r)
			}
		}
	}

	return results, nil
}

func buildOne(b config.Build, goos, goarch, dir, projectRoot string) (Result, error) {
	binName := filepath.Base(b.Binary)
	if goos == "windows" {
		binName += ".exe"
	}

	output := filepath.Join(projectRoot, dir, goos+"_"+goarch+"_"+binName)

	mainPath := b.Main
	if mainPath == "" {
		mainPath = "."
	}

	buildRoot := projectRoot
	if filepath.IsAbs(mainPath) {
		buildRoot = filepath.Dir(mainPath)
	} else if mainPath != "." {
		buildRoot = filepath.Join(projectRoot, mainPath)
	}

	goModRoot, err := findGoMod(buildRoot)
	if err != nil {
		return Result{}, fmt.Errorf("no go.mod found for %s: %w", b.ID, err)
	}

	args := []string{"build", "-o", output}

	if b.LdFlags != "" {
		args = append(args, "-ldflags", b.LdFlags)
	}

	if mainPath != "." {
		if filepath.IsAbs(mainPath) {
			args = append(args, mainPath)
		} else {
			args = append(args, ".")
		}
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = goModRoot
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building %s/%s...\n", goos, goarch)

	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("build failed for %s/%s: %w", goos, goarch, err)
	}

	return Result{
		Path:   output,
		GoOS:   goos,
		GoArch: goarch,
	}, nil
}

func findGoMod(start string) (string, error) {
	current := start
	for {
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("no go.mod found in %s or parents", start)
}
