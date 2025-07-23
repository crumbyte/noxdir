package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	tarSuffix = ".tar"
	gzSuffix  = ".gz"
)

func Pack(files []string, compress bool, outputPath string) error {
	var archiveWriter io.WriteCloser

	outputPath += tarSuffix

	if compress {
		outputPath += gzSuffix
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("archive: create output file: %w", err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	archiveWriter = outputFile

	if compress {
		gzWriter := gzip.NewWriter(outputFile)

		defer func() {
			_ = gzWriter.Close()
		}()

		archiveWriter = gzWriter
	}

	tarWriter := tar.NewWriter(archiveWriter)

	defer func() {
		_ = tarWriter.Close()
	}()

	for _, path := range files {
		if err = addToArchive(tarWriter, path); err != nil {
			return fmt.Errorf("archive: add file to archive: %w", err)
		}
	}

	return nil
}

func Unpack(archivePath string, outputPath string) error {
	var (
		reader io.Reader
		header *tar.Header
	)

	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("archive: open archive file: %w", err)
	}

	defer func() {
		_ = archiveFile.Close()
	}()

	reader = archiveFile

	compressed := strings.HasSuffix(archivePath, gzSuffix)

	if compressed {
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return fmt.Errorf("archive: create gzip reader: %w", err)
		}

		defer func() {
			_ = gzipReader.Close()
		}()

		reader = gzipReader
	}

	outputStat, err := os.Stat(outputPath)
	if (err != nil && !os.IsNotExist(err)) || (outputStat != nil && !outputStat.IsDir()) {
		return fmt.Errorf("archive: open output dir: %w", err)
	}

	if os.IsNotExist(err) {
		if err = os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("archive: create output dir: %w", err)
		}
	}

	tarReader := tar.NewReader(reader)

	for {
		if header, err = tarReader.Next(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("archive: read header: %w", err)
		}

		filePath := filepath.Join(outputPath, filepath.Base(header.Name))

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return fmt.Errorf("archive: create target file: %w", err)
		}

		if _, err = io.Copy(file, tarReader); err != nil {
			return fmt.Errorf("archive: write target file: %w", err)
		}

		_ = file.Close()
	}

	return nil
}

func addToArchive(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, stat.Name())
	if err != nil {
		return err
	}

	header.Name = path

	if err = tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)

	return err
}
