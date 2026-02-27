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

// Hash computes the NAR hash of a directory.
func Hash(path string) (string, error) {
	h := sha256.New()
	tw := tar.NewWriter(h)

	// Write version header
	if err := tw.WriteHeader(&tar.Header{
		Name:     "nix-2.11",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return "", fmt.Errorf("write nix version: %w", err)
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
		Name:     "nix-2.11",
		Mode:     0,
		Size:     0,
		Typeflag: tar.TypeDir,
	}); err != nil {
		return fmt.Errorf("write nix version: %w", err)
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

// safeExtractPath joins dest and name and ensures the result stays within dest.
func safeExtractPath(dest, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty archive entry name")
	}
	// Disallow absolute paths in the archive.
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("absolute archive entry path %q", name)
	}

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return "", fmt.Errorf("resolve dest path: %w", err)
	}

	candidate := filepath.Join(destAbs, name)
		safePath, err := safeJoinWithinRoot(dest, header.Name)
		if err != nil {
			return fmt.Errorf("invalid archive path %q: %w", header.Name, err)
		}
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}
			if err := os.MkdirAll(safePath, 0755); err != nil {
	rel, err := filepath.Rel(destAbs, targetAbs)
	if err != nil {
		return "", fmt.Errorf("rel path: %w", err)
			if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
	// If rel starts with "..", the target is outside dest.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			f, err := os.Create(safePath)
	}

	return targetAbs, nil
}

// Extract extracts the NAR archive to a directory.
func (nr *Reader) Extract(dest string) error {
	for {
			os.Chmod(safePath, header.FileInfo().Mode())
		if err == io.EOF {
			if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
		}
		if err != nil {
			if err := os.Symlink(header.Linkname, safePath); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}

// safeJoinWithinRoot joins root and candidate and ensures that the resulting path,
// after resolving any existing symlinks, stays within root.
func safeJoinWithinRoot(root, candidate string) (string, error) {
	if filepath.IsAbs(candidate) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	joined := filepath.Join(root, candidate)
	cleaned := filepath.Clean(joined)
	// Resolve symlinks to account for previously-extracted entries.
	realpath, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		// If the path (or its parents) doesn't exist yet, fall back to using the cleaned path,
		// but still enforce that it is syntactically within root.
		if !os.IsNotExist(err) {
			return "", err
		}
		realpath = cleaned
	}

	rel, err := filepath.Rel(root, realpath)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, fmt.Sprintf("..%c", os.PathSeparator)) {
		return "", fmt.Errorf("path escapes root")
	}

	return realpath, nil
}
		}

		// Skip special files
		if header.Name == "nix-2.11" || header.Name == "nix archive" {
			continue
		}

		targetPath, err := safeExtractPath(dest, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("mkdir for file: %w", err)
			}
			f, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			if _, err := io.Copy(f, nr.r); err != nil {
				f.Close()
				return fmt.Errorf("extract file: %w", err)
			}
			f.Close()
			os.Chmod(targetPath, header.FileInfo().Mode())
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("mkdir for symlink: %w", err)
			}
			os.Symlink(header.Linkname, targetPath)
		}
	}

	return nil
}
