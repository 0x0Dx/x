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
	"strings"
)
// isSafePath checks that a candidate path, interpreted under dest,
// does not escape the dest directory, taking existing symlinks into
// account by using EvalSymlinks.
func isSafePath(dest, candidate string) bool {
	if filepath.IsAbs(candidate) {
		return false
	}

	// Construct the path under dest.
	joined := filepath.Join(dest, candidate)

	// Resolve symlinks in the path. If the path (or any part of it)
	// does not exist yet, EvalSymlinks will fail; in that case we
	// fall back to using the cleaned joined path for the safety check.
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

	// Write version header
	if err := tw.WriteHeader(&tar.Header{
		Name:     "mochi-1.0",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return "", fmt.Errorf("write mochi archive: %w", err)
	}

	// Walk the directory
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

	// Write narinfodump(narVersion);
	if err := tw.WriteHeader(&tar.Header{
		Name:     "mochi-1.0",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return fmt.Errorf("write mochi archive: %w", err)
	}

	// Walk the directory
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

// Extract extracts the NAR archive to a directory.
	root, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("resolve dest: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("eval dest symlinks: %w", err)
	}

	isUnderRoot := func(p string) (string, error) {
		if filepath.IsAbs(p) {
			return "", fmt.Errorf("absolute path not allowed: %s", p)
		if !isSafePath(dest, header.Name) {
			return fmt.Errorf("unsafe path in archive entry name: %q", header.Name)
		}

		}
		joined := filepath.Join(root, p)
		real, err := filepath.EvalSymlinks(joined)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(root, real)
		if err != nil {
			return "", err
		}
		if rel == "." {
			return real, nil
		}
		if rel == ".." || len(rel) >= 3 && rel[:3] == ".."+string(os.PathSeparator) {
			return "", fmt.Errorf("path escapes root: %s", p)
		}
		return real, nil
	}

func (nr *Reader) Extract(dest string) error {
	// Ensure destination is an absolute, cleaned path.
	absDest, err := filepath.Abs(filepath.Clean(dest))
			if !isSafePath(dest, header.Linkname) {
				return fmt.Errorf("unsafe symlink target in archive: %q", header.Linkname)
			}
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	for {
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read next: %w", err)
		path, err := isUnderRoot(header.Name)
		if err != nil {
			return fmt.Errorf("invalid path in archive: %w", err)
		}

		// Skip special files
		if header.Name == "mochi-1.0" || header.Name == "mochi archive" {
			continue
		}

		// Clean and validate the header name to prevent directory traversal.
		cleanName := filepath.Clean(header.Name)
		if cleanName == "." {
			// Nothing to extract for a root-like entry.
			continue
		}

		targetPath := filepath.Join(absDest, cleanName)
		absTarget, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("invalid target path for %q: %w", header.Name, err)
		}

		// Ensure the target path is within the destination directory.
		prefix := absDest
		if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
			prefix += string(os.PathSeparator)
		}
			// Validate that the symlink target stays within the root as well.
			if _, err := isUnderRoot(header.Linkname); err != nil {
				return fmt.Errorf("invalid symlink target in archive: %w", err)
			}
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
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
			if err := os.Symlink(header.Linkname, absTarget); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
		}
	}

	return nil
}
