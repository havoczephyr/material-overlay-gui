package imagecache

import (
	"fmt"
	"os"
	"path/filepath"
)

const CacheDir = ".cache/images"

// DiskCache handles reading and writing card images to disk.
type DiskCache struct {
	dir string
}

func New() (*DiskCache, error) {
	if err := os.MkdirAll(CacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating image cache dir: %w", err)
	}
	return &DiskCache{dir: CacheDir}, nil
}

// CacheImage saves raw image bytes to disk, keyed by card ID.
func (d *DiskCache) CacheImage(cardID int, data []byte) error {
	path := filepath.Join(d.dir, fmt.Sprintf("%d_full.jpg", cardID))
	return os.WriteFile(path, data, 0o644)
}

// GetCachedImage returns the raw image bytes from disk cache, or nil if not cached.
func (d *DiskCache) GetCachedImage(cardID int) ([]byte, error) {
	path := filepath.Join(d.dir, fmt.Sprintf("%d_full.jpg", cardID))
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}

// CacheImageByName saves raw image bytes to disk, keyed by filename.
func (d *DiskCache) CacheImageByName(name string, data []byte) error {
	path := filepath.Join(d.dir, name)
	return os.WriteFile(path, data, 0o644)
}

// GetCachedImageByName returns raw image bytes from disk cache by filename, or nil if not cached.
func (d *DiskCache) GetCachedImageByName(name string) ([]byte, error) {
	path := filepath.Join(d.dir, name)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}
