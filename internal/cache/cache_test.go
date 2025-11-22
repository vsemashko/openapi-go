package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				CacheDir: t.TempDir(),
			},
			wantErr: false,
		},
		{
			name: "empty cache dir",
			config: Config{
				CacheDir: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if cache == nil {
					t.Error("NewCache() returned nil cache")
				}
				if cache.Size() != 0 {
					t.Errorf("NewCache() cache size = %d, want 0", cache.Size())
				}
			}
		})
	}
}

func TestComputeFileHash(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		wantErr    bool
		consistent bool
	}{
		{
			name:       "simple content",
			content:    "hello world",
			wantErr:    false,
			consistent: true,
		},
		{
			name:       "json content",
			content:    `{"openapi":"3.0.0","info":{"title":"Test"}}`,
			wantErr:    false,
			consistent: true,
		},
		{
			name:       "empty file",
			content:    "",
			wantErr:    false,
			consistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			hash1, err := computeFileHash(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("computeFileHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && hash1 == "" {
				t.Error("computeFileHash() returned empty hash")
			}

			// Verify consistency
			if tt.consistent {
				hash2, err := computeFileHash(filePath)
				if err != nil {
					t.Errorf("Second computeFileHash() failed: %v", err)
				}
				if hash1 != hash2 {
					t.Errorf("Hash inconsistent: %s != %s", hash1, hash2)
				}
			}
		})
	}
}

func TestComputeFileHashNonexistent(t *testing.T) {
	_, err := computeFileHash("/nonexistent/file.txt")
	if err == nil {
		t.Error("computeFileHash() should fail for nonexistent file")
	}
}

func TestCacheSet(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	outputDir := filepath.Join(tmpDir, "output")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	// Create a test spec file
	specPath := filepath.Join(tmpDir, "openapi.json")
	specContent := `{"openapi":"3.0.0"}`
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	// Set cache entry
	err = cache.Set(specPath, outputDir, "testservice", "v1.0.0")
	if err != nil {
		t.Errorf("Set() failed: %v", err)
	}

	// Verify entry was stored
	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1", cache.Size())
	}

	// Verify entry contents
	entry, exists := cache.Get(specPath)
	if !exists {
		t.Error("Get() entry not found")
	}

	if entry.ServiceName != "testservice" {
		t.Errorf("Entry.ServiceName = %s, want testservice", entry.ServiceName)
	}

	if entry.GeneratorVersion != "v1.0.0" {
		t.Errorf("Entry.GeneratorVersion = %s, want v1.0.0", entry.GeneratorVersion)
	}

	if entry.OutputPath != outputDir {
		t.Errorf("Entry.OutputPath = %s, want %s", entry.OutputPath, outputDir)
	}

	if entry.SpecHash == "" {
		t.Error("Entry.SpecHash is empty")
	}

	if entry.GeneratedAt.IsZero() {
		t.Error("Entry.GeneratedAt is zero")
	}
}

func TestCacheIsValid(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	outputDir := filepath.Join(tmpDir, "output")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	// Create a test spec file
	specPath := filepath.Join(tmpDir, "openapi.json")
	specContent := `{"openapi":"3.0.0"}`
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	tests := []struct {
		name             string
		setup            func()
		generatorVersion string
		wantValid        bool
		wantErr          bool
	}{
		{
			name: "no cache entry",
			setup: func() {
				// Do nothing - no cache entry
			},
			generatorVersion: "v1.0.0",
			wantValid:        false,
			wantErr:          false,
		},
		{
			name: "valid cache entry",
			setup: func() {
				cache.Set(specPath, outputDir, "testservice", "v1.0.0")
			},
			generatorVersion: "v1.0.0",
			wantValid:        true,
			wantErr:          false,
		},
		{
			name: "different generator version",
			setup: func() {
				cache.Set(specPath, outputDir, "testservice", "v1.0.0")
			},
			generatorVersion: "v2.0.0",
			wantValid:        false,
			wantErr:          false,
		},
		{
			name: "modified spec file",
			setup: func() {
				cache.Set(specPath, outputDir, "testservice", "v1.0.0")
				// Modify spec file
				os.WriteFile(specPath, []byte(`{"openapi":"3.1.0"}`), 0644)
			},
			generatorVersion: "v1.0.0",
			wantValid:        false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cache and spec file
			cache.Clear()
			os.WriteFile(specPath, []byte(specContent), 0644)

			// Run setup
			if tt.setup != nil {
				tt.setup()
			}

			// Check validity
			valid, err := cache.IsValid(specPath, tt.generatorVersion)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if valid != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestCacheInvalidate(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	outputDir := filepath.Join(tmpDir, "output")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	// Create a test spec file
	specPath := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	// Add cache entry
	cache.Set(specPath, outputDir, "testservice", "v1.0.0")

	if cache.Size() != 1 {
		t.Fatalf("Cache size = %d, want 1 before invalidation", cache.Size())
	}

	// Invalidate entry
	err = cache.Invalidate(specPath)
	if err != nil {
		t.Errorf("Invalidate() failed: %v", err)
	}

	if cache.Size() != 0 {
		t.Errorf("Cache size = %d, want 0 after invalidation", cache.Size())
	}

	// Verify entry is gone
	_, exists := cache.Get(specPath)
	if exists {
		t.Error("Get() found entry after invalidation")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	// Add multiple entries
	for i := 0; i < 3; i++ {
		specPath := filepath.Join(tmpDir, filepath.Join("spec", string(rune('0'+i))+".json"))
		os.MkdirAll(filepath.Dir(specPath), 0755)
		os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644)
		cache.Set(specPath, tmpDir, "service", "v1.0.0")
	}

	if cache.Size() != 3 {
		t.Fatalf("Cache size = %d, want 3 before clear", cache.Size())
	}

	// Clear cache
	err = cache.Clear()
	if err != nil {
		t.Errorf("Clear() failed: %v", err)
	}

	if cache.Size() != 0 {
		t.Errorf("Cache size = %d, want 0 after clear", cache.Size())
	}
}

func TestCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	outputDir := filepath.Join(tmpDir, "output")

	// Create first cache instance and add entry
	cache1, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	specPath := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	cache1.Set(specPath, outputDir, "testservice", "v1.0.0")

	// Create second cache instance (should load persisted data)
	cache2, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() second instance failed: %v", err)
	}

	if cache2.Size() != 1 {
		t.Errorf("Cache2 size = %d, want 1 (should load persisted data)", cache2.Size())
	}

	// Verify entry contents
	entry, exists := cache2.Get(specPath)
	if !exists {
		t.Error("Get() entry not found in second cache instance")
	}

	if entry.ServiceName != "testservice" {
		t.Errorf("Entry.ServiceName = %s, want testservice", entry.ServiceName)
	}
}

func TestCachePruneInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	// Add entries with valid and invalid spec paths
	validSpecPath := filepath.Join(tmpDir, "valid.json")
	if err := os.WriteFile(validSpecPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
		t.Fatalf("Failed to create valid spec: %v", err)
	}
	cache.Set(validSpecPath, tmpDir, "validservice", "v1.0.0")

	// Add entry for nonexistent spec
	invalidSpecPath := filepath.Join(tmpDir, "nonexistent.json")
	cache.entries[invalidSpecPath] = &Entry{
		SpecHash:         "fakehash",
		GeneratedAt:      time.Now(),
		OutputPath:       tmpDir,
		ServiceName:      "invalidservice",
		GeneratorVersion: "v1.0.0",
	}

	if cache.Size() != 2 {
		t.Fatalf("Cache size = %d, want 2 before pruning", cache.Size())
	}

	// Prune invalid entries
	pruned, err := cache.PruneInvalid()
	if err != nil {
		t.Errorf("PruneInvalid() failed: %v", err)
	}

	if pruned != 1 {
		t.Errorf("PruneInvalid() pruned %d entries, want 1", pruned)
	}

	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1 after pruning", cache.Size())
	}

	// Verify only valid entry remains
	_, exists := cache.Get(validSpecPath)
	if !exists {
		t.Error("Valid entry was pruned")
	}

	_, exists = cache.Get(invalidSpecPath)
	if exists {
		t.Error("Invalid entry was not pruned")
	}
}

func TestCacheGet(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	specPath := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644); err != nil {
		t.Fatalf("Failed to create spec file: %v", err)
	}

	// Get nonexistent entry
	_, exists := cache.Get(specPath)
	if exists {
		t.Error("Get() found nonexistent entry")
	}

	// Add entry
	cache.Set(specPath, tmpDir, "testservice", "v1.0.0")

	// Get existing entry
	entry, exists := cache.Get(specPath)
	if !exists {
		t.Error("Get() did not find existing entry")
	}

	if entry == nil {
		t.Error("Get() returned nil entry")
	}
}

func TestCacheSize(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	cache, err := NewCache(Config{CacheDir: cacheDir})
	if err != nil {
		t.Fatalf("NewCache() failed: %v", err)
	}

	if cache.Size() != 0 {
		t.Errorf("Initial cache size = %d, want 0", cache.Size())
	}

	// Add entries
	for i := 0; i < 5; i++ {
		specPath := filepath.Join(tmpDir, filepath.Join("spec", string(rune('0'+i))+".json"))
		os.MkdirAll(filepath.Dir(specPath), 0755)
		os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0644)
		cache.Set(specPath, tmpDir, "service", "v1.0.0")

		expectedSize := i + 1
		if cache.Size() != expectedSize {
			t.Errorf("Cache size = %d after %d additions, want %d", cache.Size(), expectedSize, expectedSize)
		}
	}
}
