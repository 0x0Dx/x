// Package archiver creates zip and tar.gz archives.
package archiver

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Create creates an archive from the given files.
func Create(results []string, format, output string) error {
	if format == "zip" || strings.HasSuffix(output, ".zip") {
		return createZip(results, output)
	}
	return createTarGz(results, output)
}

func createZip(files []string, output string) error {
	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", output, err)
	}
	defer func() { _ = out.Close() }()

	w := zip.NewWriter(out)
	defer func() { _ = w.Close() }()

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", f, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %w", f, err)
		}
		header.Name = filepath.Base(f)

		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create header in zip: %w", err)
		}

		reader, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", f, err)
		}
		defer func() { _ = reader.Close() }()

		if _, err := io.Copy(writer, reader); err != nil {
			return fmt.Errorf("failed to copy %s: %w", f, err)
		}
	}

	fmt.Printf("Created %s\n", output)
	return nil
}

func createTarGz(files []string, output string) error {
	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", output, err)
	}
	defer func() { _ = out.Close() }()

	gw := gzip.NewWriter(out)
	defer func() { _ = gw.Close() }()

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", f, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create header for %s: %w", f, err)
		}
		header.Name = filepath.Base(f)
		header.Method = zip.Deflate

		writer, err := zip.NewWriter(gw).CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create header in tar: %w", err)
		}

		reader, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", f, err)
		}
		defer func() { _ = reader.Close() }()

		if _, err := io.Copy(writer, reader); err != nil {
			return fmt.Errorf("failed to copy %s: %w", f, err)
		}
	}

	fmt.Printf("Created %s\n", output)
	return nil
}
