// Package eval provides expression evaluation for derivations.
package eval

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/hasher"
)

// EvalResult represents the result of evaluating an expression.
type EvalResult struct {
	Expr interface{}
	Hash hasher.Hash
}

// Environment represents build environment variables.
type Environment map[string]string

// ExecResult represents the result of executing a derivation.
type ExecResult struct {
	OutputPath string
	Hash       hasher.Hash
}

// Evaluator evaluates Nix-like expressions.
type Evaluator struct {
	DB         *db.DB
	ValuesDir  string
	LogDir     string
	SourcesDir string
}

// New creates a new Evaluator.
func New(db *db.DB, valuesDir, logDir, sourcesDir string) *Evaluator {
	return &Evaluator{
		DB:         db,
		ValuesDir:  valuesDir,
		LogDir:     logDir,
		SourcesDir: sourcesDir,
	}
}

// EvalValue evaluates an expression and returns the result.
func (e *Evaluator) EvalValue(expr interface{}) (*EvalResult, error) {
	switch x := expr.(type) {
	case string:
		// String literal
		return &EvalResult{
			Expr: x,
			Hash: hasher.FromString(x),
		}, nil
	case bool:
		return &EvalResult{
			Expr: x,
			Hash: hasher.FromString(fmt.Sprintf("%t", x)),
		}, nil
	case map[string]interface{}:
		return e.evalDerivation(x)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evalDerivation evaluates a derivation expression.
func (e *Evaluator) evalDerivation(expr map[string]interface{}) (*EvalResult, error) {
	name, _ := expr["name"].(string)
	buildPlatform, _ := expr["buildPlatform"].(string)
	system, _ := expr["system"].(string)
	builder, _ := expr["builder"].(string)
	args, _ := expr["args"].([]string)
	env, _ := expr["env"].(map[string]string)

	if name == "" {
		name = "unnamed"
	}

	if system == "" {
		system = "x86_64-linux"
	}

	// Build environment
	buildEnv := Environment{
		"out": "",
	}

	// Add user-defined env vars
	for k, v := range env {
		buildEnv[k] = v
	}

	// Execute the builder
	result, err := e.computeDerived(buildPlatform, name, builder, buildEnv, args)
	if err != nil {
		return nil, fmt.Errorf("compute derived: %w", err)
	}

	return &EvalResult{
		Expr: result.OutputPath,
		Hash: result.Hash,
	}, nil
}

// computeDerived runs a builder to create a derived value.
func (e *Evaluator) computeDerived(platform, name, builder string, env Environment, args []string) (*ExecResult, error) {
	// Check platform
	if platform != "" && platform != "x86_64-linux" && platform != getCurrentSystem() {
		return nil, fmt.Errorf("platform %s not supported, current system is %s", platform, getCurrentSystem())
	}

	// Hash the inputs to create unique output path
	inputHash := hashInputs(name, env, builder)
	outputPath := filepath.Join(e.ValuesDir, fmt.Sprintf("%s-%s", inputHash, name))

	// Check if already built
	if _, err := os.Stat(outputPath); err == nil {
		return &ExecResult{
			OutputPath: outputPath,
			Hash:       inputHash,
		}, nil
	}

	// Create temp build directory
	buildDir, err := os.MkdirTemp("", "mochii-build-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	// Set output path in environment
	env["out"] = buildDir

	// Find builder
	builderPath, err := e.findBuilder(builder)
	if err != nil {
		return nil, fmt.Errorf("find builder: %w", err)
	}

	// Create log file
	logFile := filepath.Join(e.LogDir, fmt.Sprintf("%s-%s.log", inputHash, name))
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	f, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}

	// Run builder
	cmd := exec.Command(builderPath, args...)
	cmd.Dir = buildDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	// Set environment
	goEnv := os.Environ()
	for k, v := range env {
		goEnv = append(goEnv, k+"="+v)
	}
	cmd.Env = goEnv

	if err := cmd.Run(); err != nil {
		f.Close()
		return nil, fmt.Errorf("builder failed: %w", err)
	}
	f.Close()

	// Move to final location
	if err := os.Rename(buildDir, outputPath); err != nil {
		// Copy if rename fails (cross-device)
		if err := copyDir(buildDir, outputPath); err != nil {
			return nil, fmt.Errorf("copy dir: %w", err)
		}
		os.RemoveAll(buildDir)
	}

	// Hash the result
	resultHash, err := hasher.FromFile(outputPath)
	if err != nil {
		resultHash = inputHash
	}

	// Register in database
	e.DB.Set("nfs", inputHash.String(), resultHash.String())

	return &ExecResult{
		OutputPath: outputPath,
		Hash:       resultHash,
	}, nil
}

func (e *Evaluator) findBuilder(builder string) (string, error) {
	// Try as path first
	if filepath.IsAbs(builder) {
		return builder, nil
	}

	// Try in values directory
	path := filepath.Join(e.ValuesDir, builder)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try in PATH
	path, err := exec.LookPath(builder)
	if err != nil {
		return "", fmt.Errorf("builder not found: %s", builder)
	}

	return path, nil
}

func hashInputs(name string, env Environment, builder string) hasher.Hash {
	inputStr := fmt.Sprintf("%s:%s:%s", name, builder, mapToString(env))
	return hasher.FromString(inputStr)
}

func mapToString(env Environment) string {
	var pairs []string
	for k, v := range env {
		pairs = append(pairs, k+"="+v)
	}
	return strings.Join(pairs, ",")
}

func getCurrentSystem() string {
	return "x86_64-linux"
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		return os.Chmod(dstPath, info.Mode())
	})
}
