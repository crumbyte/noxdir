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
	var reader io.Reader

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
		var gzipReader *gzip.Reader

		if gzipReader, err = gzip.NewReader(archiveFile); err != nil {
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
		if err = os.MkdirAll(outputPath, 0750); err != nil {
			return fmt.Errorf("archive: create output dir: %w", err)
		}
	}

	if err = readFromArchive(tar.NewReader(reader), outputPath); err != nil {
		return fmt.Errorf("archive: read archive: %w", err)
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

func readFromArchive(tr *tar.Reader, outputPath string) error {
	var (
		err    error
		file   *os.File
		header *tar.Header
	)

	for {
		if header, err = tr.Next(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		filePath := filepath.Join(outputPath, filepath.Base(header.Name))

		file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}

		if _, err = io.Copy(file, tr); err != nil {
			return err
		}

		_ = file.Close()
	}

	return nil
}
