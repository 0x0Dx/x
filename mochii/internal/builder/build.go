// Package builder provides package building and installation.
package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/eval"
	"github.com/0x0Dx/x/mochii/internal/hasher"
	"github.com/0x0Dx/x/mochii/internal/helper"
	"github.com/0x0Dx/x/mochii/internal/nar"
	"github.com/0x0Dx/x/mochii/internal/store"
	"github.com/0x0Dx/x/mochii/internal/values"
)

// Database table names.
const (
	DBRefs       = "refs"       // Source file locations (hash -> path)
	DBInstPkgs   = "pkginst"    // Installed packages (hash -> path)
	DBPrebuilts  = "prebuilts"  // Binary substitutes (pkg-hash -> prebuilt-hash)
	DBNetSources = "netsources" // Network URLs (hash -> URL)
)

// Package represents metadata about a package.
type Package struct {
	ID     string
	Build  string
	Run    string
	System string
	Path   string
	Hash   hasher.Hash
}

// Environment holds environment variables for builds.
type Environment map[string]string

// Builder handles package building and installation.
type Builder struct {
	DB         *db.DB
	SourcesDir string // Where source tarballs are stored
	InstallDir string // Where built packages are installed
	LogDir     string // Where build logs are stored
	ValuesDir  string // Where values are stored
	Store      *store.Store
	Values     *values.Manager
	Evaluator  *eval.Evaluator
}

// New creates a new Builder with the given directories.
func New(db *db.DB, sourcesDir, installDir, logDir string) *Builder {
	valuesDir := installDir // Values stored in install dir

	b := &Builder{
		DB:         db,
		SourcesDir: sourcesDir,
		InstallDir: installDir,
		LogDir:     logDir,
		ValuesDir:  valuesDir,
	}

	b.Store = store.New(db, sourcesDir, installDir)
	b.Values = values.New(db, valuesDir)
	b.Evaluator = eval.New(db, valuesDir, logDir, sourcesDir)

	return b
}

// GetFile returns the path to a source file given its 	hasher.
// First checks local refs, then tries to fetch from network sources.
func (b *Builder) GetFile(h hasher.Hash) (string, error) {
	for {
		// Check if we have a local reference
		if fn, ok, err := b.DB.Get(DBRefs, h.String()); err == nil && ok {
			if !helper.FileExists(fn) {
				return "", helper.Errorf("file %s does not exist", fn)
			}
			return fn, nil
		}

		// Try to find a network source
		url, ok, err := b.DB.Get(DBNetSources, h.String())
		if err != nil {
			return "", fmt.Errorf("get netsource: %w", err)
		}
		if !ok {
			return "", helper.Errorf("file with hash %s not found", h)
		}

		// Fetch from URL
		fn, err := b.fetchURL(url)
		if err != nil {
			return "", err
		}

		// Cache the reference locally
		if err := b.DB.Set(DBRefs, h.String(), fn); err != nil {
			return "", fmt.Errorf("set ref: %w", err)
		}
	}
}

// fetchURL downloads a file using wget.
func (b *Builder) fetchURL(url string) (string, error) {
	filename := helper.BaseNameOf(url)
	fullname := b.SourcesDir + "/" + filename

	if helper.FileExists(fullname) {
		return fullname, nil
	}

	fmt.Printf("fetching %s\n", url)

	if err := helper.EnsureDir(b.SourcesDir); err != nil {
		return "", fmt.Errorf("ensure sources dir: %w", err)
	}

	cmd := exec.Command("wget", "-q", "-N", url)
	cmd.Dir = b.SourcesDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("wget failed: %v %s", err, out)
	}

	if !helper.FileExists(fullname) {
		return "", fmt.Errorf("wget did not create %s", fullname)
	}

	return fullname, nil
}

// Install builds and installs a package, returning the installation path.
func (b *Builder) Install(h hasher.Hash) (string, error) {
	// Check if already installed
	path, _, err := b.DB.Get(DBInstPkgs, h.String())
	if err == nil && path != "" {
		return path, nil
	}

	// Try to evaluate the derivation expression first
	if result, err := b.evalDerivation(h); err == nil {
		fmt.Printf("evaluated derivation: %s\n", result)
		// If evaluation produced a path, use it
		if result != "" {
			return result, nil
		}
	}

	// Fall back to traditional build
	if err := b.install(h); err != nil {
		return "", fmt.Errorf("install: %w", err)
	}

	return b.InstallDir + "/" + h.String(), nil
}

// evalDerivation tries to evaluate a derivation expression by hash.
// Returns the result path if successful, or empty string if not a derivation.
func (b *Builder) evalDerivation(h hasher.Hash) (string, error) {
	// Try to load the expression from values directory
	exprPath := filepath.Join(b.ValuesDir, fmt.Sprintf("%s.expr", h))
	data, err := os.ReadFile(exprPath)
	if err != nil {
		return "", fmt.Errorf("read expression: %w", err)
	}

	var expr interface{}
	if err := json.Unmarshal(data, &expr); err != nil {
		return "", fmt.Errorf("parse expression: %w", err)
	}

	// Evaluate the expression
	result, err := b.Evaluator.EvalValue(expr)
	if err != nil {
		return "", fmt.Errorf("evaluate: %w", err)
	}

	// Check if result is a path
	switch x := result.Expr.(type) {
	case string:
		return x, nil
	case eval.ExprExternal:
		// External result - return the path from the hash
		path, err := b.Values.QueryValuePath(x.Hash)
		if err != nil {
			return "", err
		}
		return path, nil
	case eval.ExprHash:
		path, err := b.Values.QueryValuePath(x.Hash)
		if err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", fmt.Errorf("unexpected result type: %T", result.Expr)
	}
}

// install performs the actual build process:
// 1. Try prebuilt if available
// 2. Fetch and extract source
// 3. Run builder script
// 4. Copy to install directory.
func (b *Builder) install(h hasher.Hash) error {
	if err := helper.EnsureDir(b.InstallDir); err != nil {
		return fmt.Errorf("ensure install dir: %w", err)
	}

	if err := helper.EnsureDir(b.LogDir); err != nil {
		return fmt.Errorf("ensure log dir: %w", err)
	}

	fmt.Printf("building %s\n", h)

	prebuilt, ok, err := b.GetPrebuilt(h)
	//nolint:nestif
	if err == nil && ok {
		fmt.Printf("trying prebuilt %s\n", prebuilt)
		src, err := b.GetFile(hasher.Hash(prebuilt))
		if err == nil {
			fmt.Printf("using prebuilt %s\n", src)
			buildDir := b.InstallDir + "/" + h.String()
			if err := b.extractTarball(src, buildDir); err != nil {
				fmt.Fprintf(os.Stderr, "warning: prebuilt extract failed: %v\n", err)
			} else {
				if err := b.DB.Set(DBInstPkgs, h.String(), buildDir); err != nil {
					return fmt.Errorf("set inst pkgs: %w", err)
				}
				return nil
			}
		}
	}

	// Fetch source
	srcFile, err := b.GetFile(h)
	if err != nil {
		return fmt.Errorf("get file: %w", err)
	}

	// Create temporary build directory
	buildDir, err := os.MkdirTemp("", "mochii-build-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("remove temp dir: %w", err)
	}

	// Extract source
	if err := b.extractTarball(srcFile, buildDir); err != nil {
		return fmt.Errorf("extract tarball: %w", err)
	}

	// If tarball contains single directory, use that
	realBuildDir := findPackageDir(buildDir)
	if realBuildDir != "" {
		buildDir = realBuildDir
	}

	// Run builder script if present
	buildScript := buildDir + "/builder"
	if helper.FileExists(buildScript) {
		if err := b.runBuilder(buildScript, buildDir, h); err != nil {
			return fmt.Errorf("run builder: %w", err)
		}
	}

	// Move to final location
	installedDir := b.InstallDir + "/" + h.String()
	if err := os.RemoveAll(installedDir); err != nil {
		return fmt.Errorf("remove installed dir: %w", err)
	}
	if err := os.Rename(buildDir, installedDir); err != nil {
		// Fall back to copying if rename fails (cross-device)
		if err := copyDir(buildDir, installedDir); err != nil {
			return fmt.Errorf("copy dir: %w", err)
		}
		if err := os.RemoveAll(buildDir); err != nil {
			return fmt.Errorf("remove build dir: %w", err)
		}
	}

	if err := b.DB.Set(DBInstPkgs, h.String(), installedDir); err != nil {
		return fmt.Errorf("set inst pkgs: %w", err)
	}
	return nil
}

// findPackageDir checks if a directory contains exactly one subdirectory
// and returns that path (common after tarball extraction).
func findPackageDir(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) == 1 {
		return filepath.Join(dir, dirs[0])
	}

	return ""
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("rel path: %w", err)
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("readlink: %w", err)
			}
			if err := os.Symlink(link, dstPath); err != nil {
				return fmt.Errorf("symlink: %w", err)
			}
			return nil
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open src: %w", err)
		}
		if err := srcFile.Close(); err != nil {
			return fmt.Errorf("close src: %w", err)
		}
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("create dst: %w", err)
		}
		if err := dstFile.Close(); err != nil {
			return fmt.Errorf("close dst: %w", err)
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("copy: %w", err)
		}
		if err := os.Chmod(dstPath, info.Mode()); err != nil {
			return fmt.Errorf("chmod: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("filepath walk: %w", err)
	}
	return nil
}

// extractTarball extracts a tarball to a destination directory.
func (b *Builder) extractTarball(src, dest string) error {
	fmt.Printf("extracting %s to %s\n", src, dest)

	ext := filepath.Ext(src)
	var cmd *exec.Cmd

	// Support various compression formats
	if ext == ".gz" || filepath.Ext(src) == ".tgz" {
		cmd = exec.Command("tar", "xzf", src, "-C", dest)
	} else if ext == ".bz2" {
		cmd = exec.Command("tar", "xjf", src, "-C", dest)
	} else if ext == ".xz" {
		cmd = exec.Command("tar", "xJf", src, "-C", dest)
	} else if ext == ".zip" {
		cmd = exec.Command("unzip", "-q", src, "-d", dest)
	} else {
		cmd = exec.Command("tar", "xf", src, "-C", dest)
	}

	cmd.Dir = "/"
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract failed: %v %s: %w", err, out, err)
	}

	return nil
}

// runBuilder executes the builder script with appropriate environment.
func (b *Builder) runBuilder(script, dir string, h hasher.Hash) error {
	fmt.Printf("running builder %s\n", script)

	logFile := b.LogDir + "/" + h.String() + ".log"
	if err := helper.EnsureDir(b.LogDir); err != nil {
		return fmt.Errorf("ensure log dir: %w", err)
	}
	f, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close log file: %w", err)
	}

	cmd := exec.Command(script)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	// Set build environment variables
	env := os.Environ()
	env = append(env, "MOCHII_PREFIX="+dir)
	env = append(env, "MOCHII_HASH="+h.String())
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("builder failed: %w", err)
	}

	return nil
}

// Run executes an installed package with the given arguments.
func (b *Builder) Run(h hasher.Hash, args []string) error {
	path, _, err := b.DB.Get(DBInstPkgs, h.String())
	if err != nil {
		return fmt.Errorf("get inst pkgs: %w", err)
	}
	if path == "" {
		return helper.Errorf("package %s not installed", h)
	}

	bin, err := findExecutable(path)
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	execBin := filepath.Join(path, bin)
	cmd := exec.Command(execBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run cmd: %w", err)
	}
	return nil
}

// findExecutable searches for an executable file in a directory tree.
// Skips "builder" script, falls back to it if no other executable found.
func findExecutable(dir string) (string, error) {
	var search func(d string, allowBuilder bool) (string, error)
	search = func(d string, allowBuilder bool) (string, error) {
		entries, err := os.ReadDir(d)
		if err != nil {
			return "", fmt.Errorf("read dir: %w", err)
		}

		for _, e := range entries {
			if e.IsDir() {
				if name, err := search(filepath.Join(d, e.Name()), allowBuilder); err == nil {
					return name, nil
				}
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0o111 != 0 {
				// Skip builder unless no other choice
				if e.Name() == "builder" && !allowBuilder {
					continue
				}
				rel, err := filepath.Rel(d, filepath.Join(d, e.Name()))
				if err != nil {
					return e.Name(), err
				}
				return rel, nil
			}
		}

		return "", helper.Errorf("no executable found in %s", d)
	}

	name, err := search(dir, false)
	if err != nil {
		name, err = search(dir, true)
	}
	if err != nil {
		return "", err
	}
	return name, nil
}

// Delete removes an installed package.
func (b *Builder) Delete(h hasher.Hash) error {
	path, ok, err := b.DB.Get(DBInstPkgs, h.String())
	if err != nil {
		return fmt.Errorf("get inst pkgs: %w", err)
	}
	if !ok {
		return nil
	}

	// Make files writable before deletion
	if helper.DirExists(path) {
		cmd := exec.Command("chmod", "-R", "+w", path)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: chmod failed: %v\n", err)
		}

		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: rm failed: %v\n", err)
		}
	}

	if err := b.DB.Delete(DBInstPkgs, h.String()); err != nil {
		return fmt.Errorf("delete inst pkg: %w", err)
	}
	return nil
}

// ListInstalled returns all installed packages.
func (b *Builder) ListInstalled() (map[string]string, error) {
	list, err := b.DB.List(DBInstPkgs)
	if err != nil {
		return nil, fmt.Errorf("list inst pkgs: %w", err)
	}
	return list, nil
}

// RegisterFile registers a source file by computing its hash and copying to sources dir.
func (b *Builder) RegisterFile(path string) (hasher.Hash, error) {
	abs, err := helper.AbsPath(path)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}

	h, err := hasher.FromFile(abs)
	if err != nil {
		return "", fmt.Errorf("hash from file: %w", err)
	}

	srcPath := b.SourcesDir + "/" + h.String()

	_, err = os.Stat(srcPath)
	//nolint:nestif
	if err != nil && os.IsNotExist(err) {
		if err := helper.EnsureDir(b.SourcesDir); err != nil {
			return "", fmt.Errorf("ensure sources dir: %w", err)
		}
		srcFile, err := os.Open(abs)
		if err != nil {
			return "", fmt.Errorf("open source file: %w", err)
		}
		if err := srcFile.Close(); err != nil {
			return "", fmt.Errorf("close source file: %w", err)
		}
		dstFile, err := os.Create(srcPath)
		if err != nil {
			return "", fmt.Errorf("create dest file: %w", err)
		}
		if err := dstFile.Close(); err != nil {
			return "", fmt.Errorf("close dest file: %w", err)
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return "", fmt.Errorf("copy file: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("stat src path: %w", err)
	}

	if err := b.DB.Set(DBRefs, h.String(), srcPath); err != nil {
		return "", fmt.Errorf("set ref: %w", err)
	}

	return h, nil
}

// RegisterURL registers a network URL for a given 	hasher.
func (b *Builder) RegisterURL(h hasher.Hash, url string) error {
	if err := b.DB.Set(DBNetSources, h.String(), url); err != nil {
		return fmt.Errorf("set netsource: %w", err)
	}
	return nil
}

// RegisterPrebuilt registers a binary substitute for a package.
func (b *Builder) RegisterPrebuilt(pkgHash, prebuiltHash hasher.Hash) error {
	if err := b.DB.Set(DBPrebuilts, pkgHash.String(), prebuiltHash.String()); err != nil {
		return fmt.Errorf("set prebuilt: %w", err)
	}
	return nil
}

// GetPrebuilt returns the prebuilt hash for a package if available.
func (b *Builder) GetPrebuilt(pkgHash hasher.Hash) (string, bool, error) {
	val, ok, err := b.DB.Get(DBPrebuilts, pkgHash.String())
	if err != nil {
		return "", false, fmt.Errorf("get prebuilt: %w", err)
	}
	return val, ok, nil
}

// ListPrebuilts returns all registered prebuilts.
func (b *Builder) ListPrebuilts() (map[string]string, error) {
	list, err := b.DB.List(DBPrebuilts)
	if err != nil {
		return nil, fmt.Errorf("list prebuilts: %w", err)
	}
	return list, nil
}

// CollectGarbage finds packages not referenced by any profile generation.
func (b *Builder) CollectGarbage(profilePath string) ([]string, error) {
	alive := make(map[string]bool)

	// Collect all hashes from profile generations
	entries, err := os.ReadDir(profilePath)
	if err != nil {
		return nil, fmt.Errorf("read profile dir: %w", err)
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".hash") {
			data, err := os.ReadFile(filepath.Join(profilePath, name))
			if err != nil {
				continue
			}
			hashStr := strings.TrimSpace(string(data))
			alive[hashStr] = true
		}
	}

	// Find installed packages not in alive set
	installed, err := b.ListInstalled()
	if err != nil {
		return nil, fmt.Errorf("list installed: %w", err)
	}

	var toDelete []string
	for h := range installed {
		if !alive[h] {
			toDelete = append(toDelete, h)
		}
	}

	return toDelete, nil
}

// CreateNar creates a NAR archive of a package.
func (b *Builder) CreateNar(path string) (string, error) {
	hash, err := nar.Hash(path)
	if err != nil {
		return "", fmt.Errorf("nar hash: %w", err)
	}
	return hash, nil
}

// ExtractNar extracts a NAR archive.
func (b *Builder) ExtractNar(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	r := nar.NewReader(f)
	return r.Extract(dest)
}

// QueryStore performs store queries.
func (b *Builder) QueryStore(args []string) ([]store.QueryResult, error) {
	return b.Store.Query(args)
}

// VerifyStore verifies store consistency.
func (b *Builder) VerifyStore() ([]string, error) {
	return b.Store.Verify()
}

// QueryGraveyard returns unreachable store paths.
func (b *Builder) QueryGraveyard(profilePath string) ([]string, error) {
	return b.Store.QueryGraveyard(profilePath)
}

// AddValue adds a value to the store.
func (b *Builder) AddValue(path string) (hasher.Hash, error) {
	return b.Values.AddValue(path)
}

// QueryValuePath queries a value path by hash.
func (b *Builder) QueryValuePath(hash hasher.Hash) (string, error) {
	return b.Values.QueryValuePath(hash)
}

// RegisterDerivation registers a derivation expression.
// The expression is stored and can be evaluated later to build the package.
func (b *Builder) RegisterDerivation(expr interface{}) (hasher.Hash, error) {
	// Serialize the expression
	data, err := json.Marshal(expr)
	if err != nil {
		return "", fmt.Errorf("marshal expression: %w", err)
	}

	// Hash the expression
	h := hasher.FromString(string(data))

	// Store the expression
	exprPath := filepath.Join(b.ValuesDir, fmt.Sprintf("%s.expr", h))
	if _, err := os.Stat(exprPath); os.IsNotExist(err) {
		if err := helper.EnsureDir(b.ValuesDir); err != nil {
			return "", fmt.Errorf("ensure values dir: %w", err)
		}
		if err := os.WriteFile(exprPath, data, 0644); err != nil {
			return "", fmt.Errorf("write expression: %w", err)
		}
	}

	return h, nil
}

// AddDerivationValue adds a derived value (result of a build) to the store.
func (b *Builder) AddDerivationValue(path string) (hasher.Hash, error) {
	return b.Values.AddValue(path)
}
