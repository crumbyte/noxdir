package checksum

import (
	"crypto/md5"  //nolint:gosec // used on demand
	"crypto/sha1" //nolint:gosec // used on demand
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

type Encoding string

const (
	HexLower Encoding = "hex"
	HexUpper Encoding = "hex-up"
	Base64   Encoding = "base64"
)

var hashMap = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha224": sha256.New224,
	"sha256": sha256.New,
	"sha384": sha512.New384,
	"sha512": sha512.New,
}

type FileHash struct{}

func NewFileHash() *FileHash {
	return &FileHash{}
}

func (fh *FileHash) Format(rawHash []byte, e Encoding) (string, error) {
	var fmtValue string

	switch e {
	case HexLower, HexUpper:
		fmtValue = hex.EncodeToString(rawHash)

		if e == HexUpper {
			fmtValue = strings.ToUpper(fmtValue)
		}
	case Base64:
		fmtValue = base64.StdEncoding.EncodeToString(rawHash)
	default:
		return "", fmt.Errorf("unknown encoding type: %s", e)
	}

	return fmtValue, nil
}

func (fh *FileHash) Calculate(hashType string, filePath string) ([]byte, error) {
	h, err := resolveHashType(hashType)
	if err != nil {
		return nil, err
	}

	file, err := fileReader(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	if _, err = io.Copy(h, file); err != nil {
		return nil, fmt.Errorf(
			"hash: failed to calculate file hash: %s: %w", filePath, err,
		)
	}

	return h.Sum(nil), nil
}

func resolveHashType(hashTypeName string) (hash.Hash, error) {
	hashFactory, ok := hashMap[hashTypeName]
	if !ok {
		return nil, fmt.Errorf("unknown hash type: %s", hashTypeName)
	}

	return hashFactory(), nil
}

func fileReader(filePath string) (io.ReadCloser, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("hash: failed to lstat file: %w", err)
	}

	if !stat.Mode().IsRegular() {
		return nil, fmt.Errorf("hash: not a regular file: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"hash: failed to open file: %s: %w", filePath, err,
		)
	}

	return file, nil
}
