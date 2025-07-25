package archive

import (
	"archive/tar"
	"bufio"
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

type CompressionType int

const (
	Gzip CompressionType = iota
	Zstd
)

type Tar struct {
	bufferSize      int
	compressionType CompressionType
}

func NewTar(bufferSize int, ct CompressionType) *Tar {
	return &Tar{
		bufferSize:      bufferSize,
		compressionType: ct,
	}
}

func (t *Tar) PackToFile(files []string, compress bool, outputPath string) error {
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

	return t.Pack(files, compress, outputFile)
}

func (t *Tar) Pack(files []string, compress bool, w io.Writer) error {
	archiveWriter := w

	if compress {
		gzWriter := gzip.NewWriter(w)

		defer func() {
			_ = gzWriter.Close()
		}()

		archiveWriter = gzWriter
	}

	buffer := bufio.NewWriterSize(archiveWriter, t.bufferSize)
	tarWriter := tar.NewWriter(archiveWriter)

	defer func() {
		_ = buffer.Flush()
		_ = tarWriter.Close()
	}()

	for _, path := range files {
		if err := addToArchive(tarWriter, path); err != nil {
			return fmt.Errorf("archive: add file to archive: %w", err)
		}
	}

	return nil
}

func (t *Tar) UnpackFromFile(archivePath string, outputPath string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("archive: open archive file: %w", err)
	}

	defer func() {
		_ = archiveFile.Close()
	}()

	return t.Unpack(
		archiveFile, strings.HasSuffix(archivePath, gzSuffix), outputPath,
	)
}

func (t *Tar) Unpack(r io.Reader, decompress bool, outputPath string) error {
	if decompress {
		gzipReader, err := gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("archive: create gzip reader: %w", err)
		}

		defer func() {
			_ = gzipReader.Close()
		}()

		r = gzipReader
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

	buffer := bufio.NewReaderSize(r, t.bufferSize)

	if err = readFromArchive(tar.NewReader(buffer), outputPath); err != nil {
		return fmt.Errorf("archive: read archive: %w", err)
	}

	return nil
}

func addToArchive(tw *tar.Writer, path string) error {
	return filepath.Walk(path, func(entry string, fi os.FileInfo, err error) error {
		var header *tar.Header

		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(path), entry)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			header, err = tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}

			// Must end with slash to be recognized as a directory
			header.Name = filepath.ToSlash(relPath) + "/"
			header.Typeflag = tar.TypeDir

			return tw.WriteHeader(header)
		}

		header, err = tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)

		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(entry)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		_ = file.Close()

		return err
	})
}

func readFromArchive(tr *tar.Reader, outputPath string) error {
	var (
		err       error
		entryPath string
		file      *os.File
		header    *tar.Header
	)

	for {
		if header, err = tr.Next(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		entryPath, err = safeJoin(outputPath, header.Name)
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeDir {
			if err = os.MkdirAll(entryPath, 0750); err != nil {
				return err
			}

			continue
		}

		file, err = os.OpenFile(
			entryPath,
			os.O_RDWR|os.O_CREATE|os.O_EXCL,
			os.FileMode(header.Mode), //nolint:gosec // are you kidding me?
		)
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

func safeJoin(base, target string) (string, error) {
	base = filepath.Clean(base)
	fullPath := filepath.Clean(filepath.Join(base, target))

	// Ensure the final path is within the intended output directory
	if !strings.HasPrefix(fullPath, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive: illegal file path: %s", fullPath)
	}

	return fullPath, nil
}
