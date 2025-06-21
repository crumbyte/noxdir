package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/klauspost/compress/zstd"
)

// DefaultCacheDir defines a default folder name that will store cache files.
const DefaultCacheDir = "noxdir"

// ErrNoCache defines an error that may occur if the requested cache entry was
// not found.
var ErrNoCache = errors.New("cache entry not found")

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
	cachePath          string
	compressionEnabled bool
}

func NewCache(opts ...Option) (*Cache, error) {
	c := &Cache{}

	for _, opt := range opts {
		opt(c)
	}

	cachePath, err := resolveCacheDir(DefaultCacheDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	if err = os.MkdirAll(cachePath, 0750); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	c.cachePath = cachePath

	return c, nil
}

// Get retrieves a cache entry by its key and maps data to the provided target.
// The target instance must be a pointer type and support JSON unmarshalling.
// If the corresponding cache entry was not found the ErrNoCache error will be
// returned.
func (c *Cache) Get(key string, target any) error {
	var r io.Reader

	cacheFile, err := os.Open(filepath.Join(c.cachePath, c.keyHash(key)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoCache
		}

		return err
	}

	defer func() {
		_ = cacheFile.Close()
	}()

	decoder := json.NewDecoder(cacheFile)

	if c.compressionEnabled {
		if r, err = zstd.NewReader(cacheFile); err != nil {
			return err
		}

		decoder = json.NewDecoder(r)
	}

	return decoder.Decode(target)
}

// Set creates a new cache entry for specified key or overrides the existing one.
// The val instance must support JSON marshalling.
func (c *Cache) Set(key string, val any) error {
	var (
		cacheFile io.WriteCloser
		err       error
	)

	cacheFile, err = c.initEntryCache(key)
	if err != nil {
		return err
	}

	defer func() {
		_ = cacheFile.Close()
	}()

	if c.compressionEnabled {
		//TODO: consider using WithEncoderDict
		if cacheFile, err = zstd.NewWriter(cacheFile); err != nil {
			return err
		}
	}

	return json.NewEncoder(cacheFile).Encode(val)
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

func resolveCacheDir(appName string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LocalAppData")
		if len(localAppData) == 0 {
			return "", errors.New("local app data folder not found")
		}

		return filepath.Join(localAppData, appName), nil

	default:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}

		return filepath.Join(homeDir, appName), nil
	}
}
