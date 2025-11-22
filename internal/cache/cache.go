package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a cache entry for a generated client
type Entry struct {
	// SpecHash is the SHA256 hash of the OpenAPI spec file
	SpecHash string `json:"spec_hash"`
	// GeneratedAt is when the client was generated
	GeneratedAt time.Time `json:"generated_at"`
	// OutputPath is the path to the generated client directory
	OutputPath string `json:"output_path"`
	// ServiceName is the name of the service
	ServiceName string `json:"service_name"`
	// GeneratorVersion is the version of the generator used
	GeneratorVersion string `json:"generator_version"`
}

// Cache manages a hash-based cache for OpenAPI client generation
type Cache struct {
	entries  map[string]*Entry // key: spec path
	cacheDir string
}

// Config contains configuration for the cache
type Config struct {
	// CacheDir is the directory where cache metadata is stored
	CacheDir string
}

// NewCache creates a new cache instance
func NewCache(cfg Config) (*Cache, error) {
	if cfg.CacheDir == "" {
		return nil, fmt.Errorf("cache directory is required")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		entries:  make(map[string]*Entry),
		cacheDir: cfg.CacheDir,
	}

	// Load existing cache entries
	if err := cache.load(); err != nil {
		// Log warning but don't fail - we'll start with empty cache
		fmt.Printf("Warning: Failed to load cache: %v\n", err)
	}

	return cache, nil
}

// computeFileHash computes SHA256 hash of a file
func computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// IsValid checks if a cache entry is valid for the given spec file
func (c *Cache) IsValid(specPath, generatorVersion string) (bool, error) {
	// Get cached entry
	entry, exists := c.entries[specPath]
	if !exists {
		return false, nil
	}

	// Compute current hash
	currentHash, err := computeFileHash(specPath)
	if err != nil {
		return false, fmt.Errorf("failed to compute current hash: %w", err)
	}

	// Check if hash matches and generator version matches
	if entry.SpecHash != currentHash {
		return false, nil
	}

	if entry.GeneratorVersion != generatorVersion {
		return false, nil
	}

	// Verify output directory still exists
	if _, err := os.Stat(entry.OutputPath); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

// Set adds or updates a cache entry
func (c *Cache) Set(specPath, outputPath, serviceName, generatorVersion string) error {
	// Compute spec hash
	hash, err := computeFileHash(specPath)
	if err != nil {
		return fmt.Errorf("failed to compute spec hash: %w", err)
	}

	// Create entry
	entry := &Entry{
		SpecHash:         hash,
		GeneratedAt:      time.Now(),
		OutputPath:       outputPath,
		ServiceName:      serviceName,
		GeneratorVersion: generatorVersion,
	}

	// Store in memory
	c.entries[specPath] = entry

	// Persist to disk
	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}

// Get retrieves a cache entry
func (c *Cache) Get(specPath string) (*Entry, bool) {
	entry, exists := c.entries[specPath]
	return entry, exists
}

// Invalidate removes a cache entry
func (c *Cache) Invalidate(specPath string) error {
	delete(c.entries, specPath)

	// Persist changes
	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save cache after invalidation: %w", err)
	}

	return nil
}

// Clear removes all cache entries
func (c *Cache) Clear() error {
	c.entries = make(map[string]*Entry)

	// Persist changes
	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save cache after clear: %w", err)
	}

	return nil
}

// Size returns the number of cache entries
func (c *Cache) Size() int {
	return len(c.entries)
}

// cacheFilePath returns the path to the cache metadata file
func (c *Cache) cacheFilePath() string {
	return filepath.Join(c.cacheDir, "cache.json")
}

// save persists cache entries to disk
func (c *Cache) save() error {
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	cachePath := c.cacheFilePath()
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// load reads cache entries from disk
func (c *Cache) load() error {
	cachePath := c.cacheFilePath()

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// No cache file yet, start with empty cache
		return nil
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	// Unmarshal cache entries
	if err := json.Unmarshal(data, &c.entries); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return nil
}

// PruneInvalid removes cache entries for specs that no longer exist
func (c *Cache) PruneInvalid() (int, error) {
	pruned := 0

	for specPath := range c.entries {
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			delete(c.entries, specPath)
			pruned++
		}
	}

	if pruned > 0 {
		if err := c.save(); err != nil {
			return pruned, fmt.Errorf("failed to save cache after pruning: %w", err)
		}
	}

	return pruned, nil
}
