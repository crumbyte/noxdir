package cache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

const cacheDir = "cache"

// ErrNoCache defines an error that may occur if the requested cache entry was
// not found.
var ErrNoCache = errors.New("cache entry not found")

// Encoder defines a basic interface for the Encode implementations. The Cache
// uses the Encoder instance when persisting the state.
type Encoder interface {
	Encode(any) error
}

// Decoder defines a basic interface for the Decode implementations. The Cache
// uses the Decoder instance when retrieving the state.
type Decoder interface {
	Decode(any) error
}

type (
	// NewEncoder defines a function type that returns a new Encoder instance.
	// This function will be called on each cache persistent operation since
	// the io.Writer may differ.
	NewEncoder func(io.Writer) Encoder

	// NewDecoder defines a function type that returns a new Decoder instance.
	// This function will be called on each cache restoring operation since
	// the io.Reader may differ.
	NewDecoder func(io.Reader) Decoder
)

// Option defines a type for providing configuration options for Cache instance.
type Option func(*Cache)

// WithCompress enables or disables compression of cache files. By default it's
// enabled.
func WithCompress() Option {
	return func(c *Cache) {
		c.compressionEnabled = true
	}
}

// Cache provides a file cache API. It saves and restores an arbitrary data types
// which can marshaled/unmarshalled as JSON into file cache. The cache entries
// can be restored by the corresponding key. The cache files will be stored at
// configured path or default DefaultCachePath will be used.
type Cache struct {
	ei                 NewEncoder
	di                 NewDecoder
	cachePath          string
	compressionEnabled bool
}

func NewCache(ne NewEncoder, nd NewDecoder, clearCache bool, appPath string, opts ...Option) (*Cache, error) {
	c := &Cache{
		ei:        ne,
		di:        nd,
		cachePath: filepath.Join(appPath, cacheDir),
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.initCacheDir(clearCache); err != nil {
		return nil, err
	}

	return c, nil
}

// Get retrieves a cache entry by its key and maps data to the provided target.
// The target instance must be a pointer type and support JSON unmarshalling.
// If the corresponding cache entry was not found the ErrNoCache error will be
// returned.
func (c *Cache) Get(key string, target any) error {
	var (
		r                io.Reader
		compressedReader *zstd.Decoder
	)

	cacheFile, err := os.Open(filepath.Join(c.cachePath, c.keyHash(key)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoCache
		}

		return err
	}

	r = bufio.NewReaderSize(cacheFile, 5<<20)

	defer func() {
		_ = cacheFile.Close()
	}()

	if c.compressionEnabled {
		compressedReader, err = zstd.NewReader(r)
		if err != nil {
			return err
		}

		r = compressedReader
		defer compressedReader.Close()
	}

	return c.di(r).Decode(target)
}

func (c *Cache) Set(key string, val any) error {
	var (
		w                io.Writer
		compressedWriter *zstd.Encoder
	)

	cacheFile, err := c.initEntryCache(key)
	if err != nil {
		return err
	}

	bufferedWriter := bufio.NewWriterSize(cacheFile, 5<<20)

	defer func() {
		_ = bufferedWriter.Flush()
		_ = cacheFile.Close()
	}()

	w = bufferedWriter

	if c.compressionEnabled {
		compressedWriter, err = zstd.NewWriter(w)
		if err != nil {
			return err
		}

		w = compressedWriter

		defer func() {
			_ = compressedWriter.Close()
		}()
	}

	return c.ei(w).Encode(val)
}

func (c *Cache) Has(key string) bool {
	fi, err := os.Lstat(filepath.Join(c.cachePath, c.keyHash(key)))

	return err == nil && fi.Mode().IsRegular()
}

func (c *Cache) initCacheDir(clearCache bool) error {
	if clearCache {
		if err := os.RemoveAll(c.cachePath); err != nil {
			return fmt.Errorf("remove cache dir: %w", err)
		}
	}

	if err := os.MkdirAll(c.cachePath, 0750); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	return nil
}

func (c *Cache) initEntryCache(key string) (io.WriteCloser, error) {
	cachePath := filepath.Join(c.cachePath, c.keyHash(key))

	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return nil, err
	}

	return cacheFile, nil
}

func (c *Cache) keyHash(key string) string {
	if len(key) == 0 {
		return ""
	}

	h := sha256.New()
	h.Write([]byte(key))

	return hex.EncodeToString(h.Sum(nil))
}
