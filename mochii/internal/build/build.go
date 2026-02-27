package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0x0Dx/x/mochii/internal/db"
	"github.com/0x0Dx/x/mochii/internal/hash"
	"github.com/0x0Dx/x/mochii/internal/util"
)

const (
	DBRefs       = "refs"
	DBInstPkgs   = "pkginst"
	DBPrebuilts  = "prebuilts"
	DBNetSources = "netsources"
)

type Package struct {
	ID     string
	Build  string
	Run    string
	System string
	Path   string
	Hash   hash.Hash
}

type Environment map[string]string

type Builder struct {
	DB         *db.DB
	SourcesDir string
	InstallDir string
	LogDir     string
}

func New(db *db.DB, sourcesDir, installDir, logDir string) *Builder {
	return &Builder{
		DB:         db,
		SourcesDir: sourcesDir,
		InstallDir: installDir,
		LogDir:     logDir,
	}
}

func (b *Builder) GetFile(h hash.Hash) (string, error) {
	for {
		if fn, ok, err := b.DB.Get(DBRefs, h.String()); err == nil && ok {
			if !util.FileExists(fn) {
				return "", util.Errorf("file %s does not exist", fn)
			}
			return fn, nil
		}

		url, ok, err := b.DB.Get(DBNetSources, h.String())
		if err != nil {
			return "", err
		}
		if !ok {
			return "", util.Errorf("file with hash %s not found", h)
		}

		fn, err := b.fetchURL(url)
		if err != nil {
			return "", err
		}

		if err := b.DB.Set(DBRefs, h.String(), fn); err != nil {
			return "", err
		}
	}
}

func (b *Builder) fetchURL(url string) (string, error) {
	filename := util.BaseNameOf(url)
	fullname := b.SourcesDir + "/" + filename

	if util.FileExists(fullname) {
		return fullname, nil
	}

	fmt.Printf("fetching %s\n", url)

	if err := util.EnsureDir(b.SourcesDir); err != nil {
		return "", err
	}

	cmd := exec.Command("wget", "-q", "-N", url)
	cmd.Dir = b.SourcesDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("wget failed: %v %s", err, out)
	}

	if !util.FileExists(fullname) {
		return "", fmt.Errorf("wget did not create %s", fullname)
	}

	return fullname, nil
}

func (b *Builder) Install(h hash.Hash) (string, error) {
	path, _, err := b.DB.Get(DBInstPkgs, h.String())
	if err == nil && path != "" {
		return path, nil
	}

	if err := b.install(h); err != nil {
		return "", err
	}

	return b.InstallDir + "/" + h.String(), nil
}

func (b *Builder) install(h hash.Hash) error {
	if err := util.EnsureDir(b.InstallDir); err != nil {
		return err
	}

	if err := util.EnsureDir(b.LogDir); err != nil {
		return err
	}

	fmt.Printf("building %s\n", h)

	if prebuilt, ok, err := b.GetPrebuilt(h); err == nil && ok {
		fmt.Printf("trying prebuilt %s\n", prebuilt)
		src, err := b.GetFile(hash.Hash(prebuilt))
		if err == nil {
			fmt.Printf("using prebuilt %s\n", src)
			buildDir := b.InstallDir + "/" + h.String()
			if err := b.extractTarball(src, buildDir); err != nil {
				fmt.Fprintf(os.Stderr, "warning: prebuilt extract failed: %v\n", err)
			} else {
				return b.DB.Set(DBInstPkgs, h.String(), buildDir)
			}
		}
	}

	srcFile, err := b.GetFile(h)
	if err != nil {
		return err
	}

	buildDir, err := os.MkdirTemp("", "mochii-build-")
	if err != nil {
		return err
	}
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}

	if err := b.extractTarball(srcFile, buildDir); err != nil {
		return err
	}

	realBuildDir := findPackageDir(buildDir)
	if realBuildDir != "" {
		buildDir = realBuildDir
	}

	buildScript := buildDir + "/builder"
	if util.FileExists(buildScript) {
		if err := b.runBuilder(buildScript, buildDir, h); err != nil {
			return err
		}
	}

	installedDir := b.InstallDir + "/" + h.String()
	if err := os.RemoveAll(installedDir); err != nil {
		return err
	}
	if err := os.Rename(buildDir, installedDir); err != nil {
		if err := copyDir(buildDir, installedDir); err != nil {
			return err
		}
		if err := os.RemoveAll(buildDir); err != nil {
			return err
		}
	}

	return b.DB.Set(DBInstPkgs, h.String(), installedDir)
}

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
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, dstPath)
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		if err := srcFile.Close(); err != nil {
			return err
		}
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		if err := dstFile.Close(); err != nil {
			return err
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		return os.Chmod(dstPath, info.Mode())
	})
}

func (b *Builder) extractTarball(src, dest string) error {
	fmt.Printf("extracting %s to %s\n", src, dest)

	ext := filepath.Ext(src)
	var cmd *exec.Cmd

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
		return fmt.Errorf("extract failed: %v %s", err, out)
	}

	return nil
}

func (b *Builder) runBuilder(script, dir string, h hash.Hash) error {
	fmt.Printf("running builder %s\n", script)

	logFile := b.LogDir + "/" + h.String() + ".log"
	if err := util.EnsureDir(b.LogDir); err != nil {
		return err
	}
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	cmd := exec.Command(script)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	env := os.Environ()
	env = append(env, "MOCHII_PREFIX="+dir)
	env = append(env, "MOCHII_HASH="+h.String())
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("builder failed: %v", err)
	}

	return nil
}

func (b *Builder) Run(h hash.Hash, args []string) error {
	path, _, err := b.DB.Get(DBInstPkgs, h.String())
	if err != nil || path == "" {
		return util.Errorf("package %s not installed", h)
	}

	bin, err := findExecutable(path)
	if err != nil {
		return err
	}

	execBin := filepath.Join(path, bin)
	cmd := exec.Command(execBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path

	return cmd.Run()
}

func findExecutable(dir string) (string, error) {
	var search func(d string, allowBuilder bool) (string, error)
	search = func(d string, allowBuilder bool) (string, error) {
		entries, err := os.ReadDir(d)
		if err != nil {
			return "", err
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
			if info.Mode()&0111 != 0 {
				if e.Name() == "builder" && !allowBuilder {
					continue
				}
				rel, err := filepath.Rel(d, filepath.Join(d, e.Name()))
				if err != nil {
					return e.Name(), nil
				}
				return rel, nil
			}
		}

		return "", util.Errorf("no executable found in %s", d)
	}

	name, err := search(dir, false)
	if err != nil {
		name, err = search(dir, true)
	}
	return name, err
}

func (b *Builder) Delete(h hash.Hash) error {
	path, ok, err := b.DB.Get(DBInstPkgs, h.String())
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if util.DirExists(path) {
		cmd := exec.Command("chmod", "-R", "+w", path)
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: chmod failed: %v\n", err)
		}

		if err := os.RemoveAll(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: rm failed: %v\n", err)
		}
	}

	return b.DB.Delete(DBInstPkgs, h.String())
}

func (b *Builder) ListInstalled() (map[string]string, error) {
	return b.DB.List(DBInstPkgs)
}

func (b *Builder) RegisterFile(path string) (hash.Hash, error) {
	abs, err := util.AbsPath(path)
	if err != nil {
		return "", err
	}

	h, err := hash.FromFile(abs)
	if err != nil {
		return "", err
	}

	srcPath := b.SourcesDir + "/" + h.String()
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		if err := util.EnsureDir(b.SourcesDir); err != nil {
			return "", err
		}
		srcFile, err := os.Open(abs)
		if err != nil {
			return "", err
		}
		if err := srcFile.Close(); err != nil {
			return "", err
		}
		dstFile, err := os.Create(srcPath)
		if err != nil {
			return "", err
		}
		if err := dstFile.Close(); err != nil {
			return "", err
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return "", err
		}
	}

	if err := b.DB.Set(DBRefs, h.String(), srcPath); err != nil {
		return "", err
	}

	return h, nil
}

func (b *Builder) RegisterURL(h hash.Hash, url string) error {
	return b.DB.Set(DBNetSources, h.String(), url)
}

func (b *Builder) RegisterPrebuilt(pkgHash, prebuiltHash hash.Hash) error {
	return b.DB.Set(DBPrebuilts, pkgHash.String(), prebuiltHash.String())
}

func (b *Builder) GetPrebuilt(pkgHash hash.Hash) (string, bool, error) {
	return b.DB.Get(DBPrebuilts, pkgHash.String())
}

func (b *Builder) ListPrebuilts() (map[string]string, error) {
	return b.DB.List(DBPrebuilts)
}

func (b *Builder) CollectGarbage(profilePath string) ([]string, error) {
	alive := make(map[string]bool)

	entries, err := os.ReadDir(profilePath)
	if err != nil {
		return nil, err
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

	installed, err := b.ListInstalled()
	if err != nil {
		return nil, err
	}

	var toDelete []string
	for h := range installed {
		if !alive[h] {
			toDelete = append(toDelete, h)
		}
	}

	return toDelete, nil
}
