package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetRepositoryRoot(t *testing.T) {
	root := GetRepositoryRoot()

	// Should not be empty
	if root == "" {
		t.Fatal("GetRepositoryRoot returned empty string")
	}

	// Should be an absolute path
	if !filepath.IsAbs(root) {
		t.Errorf("GetRepositoryRoot returned relative path: %s", root)
	}

	// Should contain go.mod
	goModPath := filepath.Join(root, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("Repository root does not contain go.mod: %s", root)
	}
}

func TestGetOgenConfigPath(t *testing.T) {
	path := GetOgenConfigPath()

	// Should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("GetOgenConfigPath returned relative path: %s", path)
	}

	// Should end with ogen.yml
	if filepath.Base(path) != "ogen.yml" {
		t.Errorf("GetOgenConfigPath does not end with ogen.yml: %s", path)
	}

	// Should exist (in this repository)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("ogen.yml not found at expected path: %s", path)
	}
}

func TestGetTemplatesDir(t *testing.T) {
	dir := GetTemplatesDir()

	// Should be absolute
	if !filepath.IsAbs(dir) {
		t.Errorf("GetTemplatesDir returned relative path: %s", dir)
	}

	// Should end with templates
	if filepath.Base(dir) != "templates" {
		t.Errorf("GetTemplatesDir does not end with 'templates': %s", dir)
	}
}

func TestGetInternalClientTemplatePath(t *testing.T) {
	path := GetInternalClientTemplatePath()

	// Should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("GetInternalClientTemplatePath returned relative path: %s", path)
	}

	// Should end with internal_client.tmpl
	if filepath.Base(path) != "internal_client.tmpl" {
		t.Errorf("GetInternalClientTemplatePath does not end with internal_client.tmpl: %s", path)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	// Should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath returned relative path: %s", path)
	}

	// Should end with application.yml
	if filepath.Base(path) != "application.yml" {
		t.Errorf("GetConfigPath does not end with application.yml: %s", path)
	}

	// Should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("application.yml not found at expected path: %s", path)
	}
}

func TestGetResourcesDir(t *testing.T) {
	dir := GetResourcesDir()

	// Should be absolute
	if !filepath.IsAbs(dir) {
		t.Errorf("GetResourcesDir returned relative path: %s", dir)
	}

	// Should end with resources
	if filepath.Base(dir) != "resources" {
		t.Errorf("GetResourcesDir does not end with 'resources': %s", dir)
	}
}

func TestEnsurePathExists(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "existing file",
			path:    GetConfigPath(), // We know this exists
			wantErr: false,
		},
		{
			name:    "existing directory",
			path:    GetResourcesDir(), // We know this exists
			wantErr: false,
		},
		{
			name:    "nonexistent path",
			path:    "/nonexistent/path/to/nowhere",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsurePathExists(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePathExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("EnsurePathExists() unexpected error: %v", err)
			}
		})
	}
}

func TestEnsureDirectoryWritable(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "new directory",
			setup: func() string {
				return filepath.Join(t.TempDir(), "test_dir")
			},
			wantErr: false,
		},
		{
			name: "existing writable directory",
			setup: func() string {
				return t.TempDir()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup()
			err := EnsureDirectoryWritable(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDirectoryWritable() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If no error, verify directory exists and is writable
			if err == nil {
				// Try to create a test file
				testFile := filepath.Join(dir, "test.txt")
				f, err := os.Create(testFile)
				if err != nil {
					t.Errorf("Directory not writable after EnsureDirectoryWritable: %v", err)
				} else {
					f.Close()
					os.Remove(testFile)
				}
			}
		})
	}
}

func TestMakeAbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAbs  bool
		wantBase string
	}{
		{
			name:     "relative path",
			input:    "resources/application.yml",
			wantAbs:  true,
			wantBase: "application.yml",
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path/file.txt",
			wantAbs:  true,
			wantBase: "file.txt",
		},
		{
			name:     "current directory",
			input:    ".",
			wantAbs:  true,
			wantBase: filepath.Base(GetRepositoryRoot()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeAbsolutePath(tt.input)

			// Check if absolute
			if filepath.IsAbs(result) != tt.wantAbs {
				t.Errorf("MakeAbsolutePath(%q) = %q, isAbs = %v, want %v",
					tt.input, result, filepath.IsAbs(result), tt.wantAbs)
			}

			// Check base name
			if filepath.Base(result) != tt.wantBase {
				t.Errorf("MakeAbsolutePath(%q) base = %q, want %q",
					tt.input, filepath.Base(result), tt.wantBase)
			}
		})
	}
}

func TestMakeAbsolutePathConsistency(t *testing.T) {
	// Same relative path should always produce same absolute path
	input := "test/path"
	result1 := MakeAbsolutePath(input)
	result2 := MakeAbsolutePath(input)

	if result1 != result2 {
		t.Errorf("MakeAbsolutePath not consistent: %q != %q", result1, result2)
	}
}
