// Package nar provides Nix ARchive (NAR) format support.
package nar

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// isSafePath checks that a candidate path, interpreted under dest,
// does not escape the dest directory, taking existing symlinks into
// account by using EvalSymlinks.
func isSafePath(dest, candidate string) bool {
	if filepath.IsAbs(candidate) {
		return false
	}

	joined := filepath.Join(dest, candidate)

	realpath, err := filepath.EvalSymlinks(joined)
	if err != nil {
		realpath = filepath.Clean(joined)
	}

	rel, err := filepath.Rel(dest, realpath)
	if err != nil {
		return false
	}

	rel = filepath.Clean(rel)
	if rel == "." {
		return true
	}

	return !filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Hash computes the NAR hash of a directory.
func Hash(path string) (string, error) {
	h := sha256.New()
	tw := tar.NewWriter(h)

	if err := tw.WriteHeader(&tar.Header{
		Name:     "mochi-1.0",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return "", fmt.Errorf("write mochi archive: %w", err)
	}

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create header for %s: %w", p, err)
		}
		header.Name = rel

		if info.IsDir() {
			header.Name += "/"
		} else if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			header.Linkname = link
			header.Typeflag = tar.TypeSymlink
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header: %w", err)
		}

		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("copy file: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if err := tw.Close(); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Writer writes NAR archives.
type Writer struct {
	w io.Writer
}

// NewWriter creates a new NAR writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteDir writes a directory to the NAR archive.
func (nw *Writer) WriteDir(path string) error {
	tw := tar.NewWriter(nw.w)

	if err := tw.WriteHeader(&tar.Header{
		Name:     "mochi-1.0",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return fmt.Errorf("write mochi archive: %w", err)
	}

	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create header for %s: %w", p, err)
		}
		header.Name = rel

		if info.IsDir() {
			header.Name += "/"
		} else if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			header.Linkname = link
			header.Typeflag = tar.TypeSymlink
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header: %w", err)
		}

		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("copy file: %w", err)
			}
		}

		return nil
	})
}

// Reader reads NAR archives.
type Reader struct {
	r *tar.Reader
}

// NewReader creates a new NAR reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: tar.NewReader(r)}
}

func (nr *Reader) Extract(dest string) error {
	absDest, err := filepath.Abs(filepath.Clean(dest))
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	for {
		header, err := nr.r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read next: %w", err)
		}

		if header.Name == "mochi-1.0" || header.Name == "mochi archive" {
			continue
		}

		cleanName := filepath.Clean(header.Name)
		if cleanName == "." {
			continue
		}
		if filepath.IsAbs(cleanName) || strings.Contains(cleanName, "..") {
			return fmt.Errorf("archive entry %q has unsafe path %q", header.Name, cleanName)
		}

		targetPath := filepath.Join(absDest, cleanName)
		absTarget, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("invalid target path for %q: %w", header.Name, err)
		}

		prefix := absDest
		if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
			prefix += string(os.PathSeparator)
		}

		checkPath := absTarget
		if !strings.HasSuffix(checkPath, string(os.PathSeparator)) && header.Typeflag == tar.TypeDir {
			checkPath += string(os.PathSeparator)
		}
		if absTarget != absDest && !strings.HasPrefix(checkPath, prefix) {
			return fmt.Errorf("archive entry %q would be extracted outside destination", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(absTarget, 0755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(absTarget), 0755); err != nil {
				return fmt.Errorf("mkdir for file: %w", err)
			}
			f, err := os.Create(absTarget)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			if _, err := io.Copy(f, nr.r); err != nil {
				f.Close()
				return fmt.Errorf("extract file: %w", err)
			}
			f.Close()
			os.Chmod(absTarget, header.FileInfo().Mode())
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(absTarget), 0755); err != nil {
				return fmt.Errorf("mkdir for symlink: %w", err)
			}
			linkTarget := header.Linkname
			if filepath.IsAbs(linkTarget) {
				return fmt.Errorf("absolute symlink target %q for %q is not allowed", linkTarget, header.Name)
			}
			linkDir := filepath.Dir(absTarget)
			candidatePath := filepath.Clean(filepath.Join(linkDir, linkTarget))
			resolvedTarget, err := filepath.EvalSymlinks(candidatePath)
			if err == nil {
				linkPrefix := absDest
				if !strings.HasSuffix(linkPrefix, string(os.PathSeparator)) {
					linkPrefix += string(os.PathSeparator)
				}
				if resolvedTarget != absDest && !strings.HasPrefix(resolvedTarget, linkPrefix) {
					return fmt.Errorf("symlink %q with target %q would resolve outside destination", header.Name, linkTarget)
				}
			}
			if err := os.Symlink(linkTarget, absTarget); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
		}
	}

	return nil
}
