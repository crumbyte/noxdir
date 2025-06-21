package cache

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

	if err = os.MkdirAll(cachePath, 0600); err != nil {
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
	var rc io.ReadCloser

	rc, err := os.Open(filepath.Join(c.cachePath, c.keyHash(key)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoCache
		}

		return err
	}

	defer func(rc io.ReadCloser) {
		_ = rc.Close()
	}(rc)

	if c.compressionEnabled {
		if rc, err = gzip.NewReader(rc); err != nil {
			return err
		}
	}

	if err = json.NewDecoder(rc).Decode(target); err != nil {
		return err
	}

	return nil
}

// Set creates a new cache entry for specified key or overrides the existing one.
// The val instance must support JSON marshalling.
func (c *Cache) Set(key string, val any) error {
	cacheFile, err := c.initEntryCache(key)
	if err != nil {
		return err
	}

	defer func(f io.WriteCloser) {
		_ = f.Close()
	}(cacheFile)

	rawData, err := json.Marshal(val)
	if err != nil {
		return err
	}

	if err = c.compress(cacheFile, rawData); err != nil {
		return err
	}

	return nil
}

func (c *Cache) compress(w io.WriteCloser, data []byte) error {
	if c.compressionEnabled {
		w = gzip.NewWriter(w)

		defer func(wc io.WriteCloser) {
			_ = wc.Close()
		}(w)
	}

	if _, err := w.Write(data); err != nil {
		return err
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

		return filepath.Join(homeDir, ".cache", appName), nil
	}
}
