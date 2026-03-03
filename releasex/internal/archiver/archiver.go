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

func Create(results []string, format, output string) error {
	if format == "zip" || strings.HasSuffix(output, ".zip") {
		return createZip(results, output)
	}
	return createTarGz(results, output)
}

func createZip(files []string, output string) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Base(f)

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		reader, err := os.Open(f)
		if err != nil {
			return err
		}
		defer reader.Close()

		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
	}

	fmt.Printf("Created %s\n", output)
	return nil
}

func createTarGz(files []string, output string) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Base(f)
		header.Method = zip.Deflate

		writer, err := zip.NewWriter(gw).CreateHeader(header)
		if err != nil {
			return err
		}

		reader, err := os.Open(f)
		if err != nil {
			return err
		}
		defer reader.Close()

		if _, err := io.Copy(writer, reader); err != nil {
			return err
		}
	}

	fmt.Printf("Created %s\n", output)
	return nil
}
