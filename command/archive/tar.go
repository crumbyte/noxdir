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

	"github.com/klauspost/compress/zstd"
)

const tarSuffix = ".tar"

// CompressionType defines a custom type the represents a compression type value.
type CompressionType int

// FromString resolves and returns the CompressionType value from the provided
// string. If string value is unknows or invalid the NoCompression will be used.
func (ct CompressionType) FromString(s string) CompressionType {
	switch s {
	case "gzip", "gz":
		return Gzip
	case "zst", "zstd":
		return Zstd
	default:
		return NoCompression
	}
}

// Extension returns the archive name extension representing the compression type.
func (ct CompressionType) Extension() string {
	switch ct {
	case Gzip:
		return ".gz"
	case Zstd:
		return ".zst"
	default:
		return ""
	}
}

func (ct CompressionType) Writer(w io.Writer) (io.WriteCloser, error) {
	switch ct {
	case Gzip:
		return gzip.NewWriter(w), nil
	case Zstd:
		zstdWriter, err := zstd.NewWriter(w)
		if err != nil {
			return nil, fmt.Errorf("archive: create zstd writer: %w", err)
		}

		return zstdWriter, nil
	default:
		return nil, nil
	}
}

func (ct CompressionType) Reader(r io.Reader) (io.Reader, error) {
	switch ct {
	case Gzip:
		gzReader, err := gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("archive: create gzip reader: %w", err)
		}

		return gzReader, nil
	case Zstd:
		ztsReader, err := zstd.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("archive: create zstd reader: %w", err)
		}

		return ztsReader, nil
	default:
		return nil, nil
	}
}

const (
	NoCompression CompressionType = iota
	Gzip
	Zstd
)

// DefaultBufferSize defines a default buffer size for reading and writing the
// archives.
const DefaultBufferSize = 5 << 20

type Tar struct {
	bufferSize      int
	compressionType CompressionType
}

func NewTar(bufferSize int, ct CompressionType) *Tar {
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	return &Tar{
		bufferSize:      bufferSize,
		compressionType: ct,
	}
}

func (t *Tar) PackToFile(files []string, outputPath string) error {
	outputPath += tarSuffix
	outputPath += t.compressionType.Extension()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("archive: create output file: %w", err)
	}

	if err = t.Pack(files, outputFile); err != nil {
		_ = outputFile.Close()

		return errors.Join(err, os.Remove(outputPath))
	}

	return outputFile.Close()
}

func (t *Tar) Pack(files []string, w io.Writer) error {
	archiveWriter := w

	compressionWriter, err := t.compressionType.Writer(archiveWriter)
	if err != nil {
		return err
	}

	if compressionWriter != nil {
		archiveWriter = compressionWriter

		defer func() {
			_ = compressionWriter.Close()
		}()
	}

	buffer := bufio.NewWriterSize(archiveWriter, t.bufferSize)
	tarWriter := tar.NewWriter(archiveWriter)

	defer func() {
		_ = buffer.Flush()
		_ = tarWriter.Close()
	}()

	for _, path := range files {
		if err = addToArchive(tarWriter, path); err != nil {
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

	outputStat, err := os.Stat(outputPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("archive: stat output file: %w", err)
	}

	if outputStat != nil && !outputStat.IsDir() {
		return fmt.Errorf("archive: output path not a directory: %s", outputPath)
	}

	if os.IsNotExist(err) {
		if err = os.MkdirAll(outputPath, 0750); err != nil {
			return fmt.Errorf("archive: create output dir: %w", err)
		}
	}

	return t.Unpack(archiveFile, outputPath)
}

func (t *Tar) Unpack(r io.Reader, outputPath string) error {
	compressionReader, err := t.compressionType.Reader(r)
	if err != nil {
		return err
	}

	if compressionReader != nil {
		r = compressionReader
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
			_ = file.Close()

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
