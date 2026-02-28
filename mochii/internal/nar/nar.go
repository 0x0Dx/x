// Package nar provides Nix ARchive (NAR) format support.
// NAR (Nix ARchive) is a serialization format for directory trees.
package nar

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const ArchiveVersion1 = "mochi-archive-1"

type DumpSink interface {
	Write(data []byte) (int, error)
}

type RestoreSource interface {
	Read(data []byte) (int, error)
}

type hashSink struct {
	h hash.Hash
}

func (s *hashSink) Write(data []byte) (int, error) {
	s.h.Write(data)
	return len(data), nil
}

func writePadding(len int, w DumpSink) {
	if pad := len % 8; pad != 0 {
		w.Write(make([]byte, 8-pad))
	}
}

func writeInt(n int, w DumpSink) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(n))
	w.Write(buf[:])
}

func writeString(s string, w DumpSink) {
	writeInt(len(s), w)
	w.Write([]byte(s))
	writePadding(len(s), w)
}

func Hash(path string) (string, error) {
	h := sha256.New()
	sink := &hashSink{h: h}
	if err := dumpPath(path, sink); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func dumpPath(path string, sink DumpSink) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat %s: %w", path, err)
	}

	writeString("(", sink)
	writeString("type", sink)

	if info.IsDir() {
		writeString("directory", sink)
		writeString("entries", sink)
		dumpEntries(path, sink)
	} else if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(path)
		if err != nil {
			return fmt.Errorf("readlink %s: %w", path, err)
		}
		writeString("symlink", sink)
		writeString("target", sink)
		writeString(link, sink)
	} else {
		writeString("regular", sink)
		writeString("contents", sink)
		dumpContents(path, info.Size(), sink)
	}

	writeString(")", sink)
	return nil
}

func dumpEntries(path string, sink DumpSink) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", path, err)
	}

	var names []string
	for _, e := range entries {
		if e.Name() == "." || e.Name() == ".." {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	writeString("(", sink)
	for i, name := range names {
		if i > 0 {
		}
		writeString("entry", sink)
		writeString("(", sink)
		writeString("name", sink)
		writeString(name, sink)
		writeString("node", sink)
		dumpPath(filepath.Join(path, name), sink)
		writeString(")", sink)
	}
	writeString(")", sink)

	return nil
}

func dumpContents(path string, size int64, sink DumpSink) error {
	writeInt(int(size), sink)

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file %s: %w", path, err)
	}
	defer f.Close()

	buf := make([]byte, 65536)
	var total int64
	for {
		n, err := f.Read(buf)
		if n > 0 {
			sink.Write(buf[:n])
			total += int64(n)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read file %s: %w", path, err)
		}
	}

	if total != size {
		return fmt.Errorf("file changed while reading: %s", path)
	}

	writePadding(int(size), sink)
	return nil
}

type writerSink struct {
	w io.Writer
}

func (s *writerSink) Write(data []byte) (int, error) {
	return s.w.Write(data)
}

func CreateNar(w io.Writer, path string) error {
	sink := &writerSink{w: w}
	writeString(ArchiveVersion1, sink)
	return dumpPath(path, sink)
}

func CreateNarHash(path string) (string, error) {
	h := sha256.New()
	sink := &hashSink{h: h}
	writeString(ArchiveVersion1, sink)
	if err := dumpPath(path, sink); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type readerSource struct {
	r io.Reader
}

func (s *readerSource) Read(data []byte) (int, error) {
	return s.r.Read(data)
}

func readPadding(len int, source RestoreSource) error {
	if pad := len % 8; pad != 0 {
		buf := make([]byte, 8-pad)
		n, err := source.Read(buf)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			if buf[i] != 0 {
				return fmt.Errorf("non-zero padding")
			}
		}
	}
	return nil
}

func readInt(source RestoreSource) (int, error) {
	buf := make([]byte, 8)
	n, err := source.Read(buf)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("not enough bytes for int")
	}
	return int(binary.LittleEndian.Uint64(buf)), nil
}

func readString(source RestoreSource) (string, error) {
	len, err := readInt(source)
	if err != nil {
		return "", err
	}
	buf := make([]byte, len)
	n, err := source.Read(buf)
	if err != nil {
		return "", err
	}
	if n != len {
		return "", fmt.Errorf("not enough bytes for string")
	}
	if err := readPadding(len, source); err != nil {
		return "", err
	}
	return string(buf), nil
}

func badArchive(s string) error {
	return fmt.Errorf("bad archive: %s", s)
}

func skipGeneric(source RestoreSource) error {
	s, err := readString(source)
	if err != nil {
		return err
	}
	if s == "(" {
		for {
			s, err := readString(source)
			if err != nil {
				return err
			}
			if s == ")" {
				break
			}
			if err := skipGeneric(source); err != nil {
				return err
			}
		}
	}
	return nil
}

func restoreEntry(path string, source RestoreSource) error {
	var name string
	for {
		s, err := readString(source)
		if err != nil {
			return err
		}
		if s == ")" {
			break
		} else if s == "name" {
			name, err = readString(source)
			if err != nil {
				return err
			}
		} else if s == "node" {
			if name == "" {
				return badArchive("entry name missing")
			}
			if err := restore(path+"/"+name, source); err != nil {
				return err
			}
		} else {
			return badArchive("unknown field: " + s)
		}
	}
	return nil
}

func restoreContents(f *os.File, path string, source RestoreSource) error {
	size, err := readInt(source)
	if err != nil {
		return err
	}
	buf := make([]byte, 65536)
	left := size
	for left > 0 {
		n := len(buf)
		if n > left {
			n = left
		}
		nr, err := source.Read(buf[:n])
		if err != nil {
			return err
		}
		nw, err := f.Write(buf[:nr])
		if err != nil {
			return err
		}
		left -= nw
	}
	return readPadding(size, source)
}

func restore(path string, source RestoreSource) error {
	s, err := readString(source)
	if err != nil {
		return err
	}
	if s != "(" {
		return badArchive("expected open tag")
	}

	var fileType string
	var f *os.File
	var targetName string

	for {
		s, err := readString(source)
		if err != nil {
			return err
		}
		if s == ")" {
			break
		} else if s == "type" {
			fileType, err = readString(source)
			if err != nil {
				return err
			}
			if fileType == "regular" {
				f, err = os.Create(path)
				if err != nil {
					return fmt.Errorf("creating file %s: %w", path, err)
				}
			} else if fileType == "directory" {
				if err := os.MkdirAll(path, 0777); err != nil {
					return fmt.Errorf("creating directory %s: %w", path, err)
				}
			} else if fileType == "symlink" {
			} else {
				return badArchive("unknown file type: " + fileType)
			}
		} else if s == "contents" && fileType == "regular" {
			if err := restoreContents(f, path, source); err != nil {
				return err
			}
		} else if s == "entry" && fileType == "directory" {
			if err := restoreEntry(path, source); err != nil {
				return err
			}
		} else if s == "target" && fileType == "symlink" {
			targetName, err = readString(source)
			if err != nil {
				return err
			}
		} else {
			if err := skipGeneric(source); err != nil {
				return err
			}
		}
	}

	if f != nil {
		f.Close()
	}
	if fileType == "symlink" && targetName != "" {
		if err := os.Symlink(targetName, path); err != nil {
			return fmt.Errorf("creating symlink %s: %w", path, err)
		}
	}

	return nil
}

func RestoreNar(r io.Reader, path string) error {
	source := &readerSource{r: r}
	version, err := readString(source)
	if err != nil {
		return err
	}
	if version != ArchiveVersion1 {
		return badArchive("expected " + ArchiveVersion1)
	}
	return restore(path, source)
}

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Extract(dest string) error {
	return RestoreNar(r.r, dest)
}
