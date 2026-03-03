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

func Build(cfg *config.Config, dir string) ([]Result, error) {
	var results []Result

	for _, b := range cfg.Builds {
		for _, goos := range b.GoOS {
			for _, goarch := range b.GoArch {
				r, err := buildOne(b, goos, goarch, dir)
				if err != nil {
					return nil, err
				}
				results = append(results, r)
			}
		}
	}

	return results, nil
}

func buildOne(b config.Build, goos, goarch, dir string) (Result, error) {
	binName := b.Binary
	if goos == "windows" {
		binName += ".exe"
	}

	output := filepath.Join(dir, goos+"_"+goarch, binName)

	args := []string{"build", "-o", output}
	if b.Main != "" {
		args = append(args, b.Main)
	}

	if b.LdFlags != "" {
		args = append(args, "-ldflags", b.LdFlags)
	}

	cmd := exec.Command("go", args...)
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
