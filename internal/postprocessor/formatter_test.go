package postprocessor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFormatterProcessor(t *testing.T) {
	tests := []struct {
		name     string
		simplify bool
	}{
		{
			name:     "with simplify",
			simplify: true,
		},
		{
			name:     "without simplify",
			simplify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewFormatterProcessor(tt.simplify)

			if processor == nil {
				t.Fatal("NewFormatterProcessor() returned nil")
			}

			if processor.simplify != tt.simplify {
				t.Errorf("simplify = %v, want %v", processor.simplify, tt.simplify)
			}
		})
	}
}

func TestFormatterProcessorName(t *testing.T) {
	processor := NewFormatterProcessor(false)
	name := processor.Name()

	if name != "GoFormatter" {
		t.Errorf("Name() = %q, want %q", name, "GoFormatter")
	}
}

func TestFormatterProcessorFindGoFiles(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(string) error
		expectedCount int
	}{
		{
			name: "single go file",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.go"), []byte("package test"), 0644)
			},
			expectedCount: 1,
		},
		{
			name: "multiple go files",
			setup: func(dir string) error {
				os.WriteFile(filepath.Join(dir, "file1.go"), []byte("package test"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.go"), []byte("package test"), 0644)
				os.WriteFile(filepath.Join(dir, "file3.go"), []byte("package test"), 0644)
				return nil
			},
			expectedCount: 3,
		},
		{
			name: "go files in subdirectory",
			setup: func(dir string) error {
				subdir := filepath.Join(dir, "subdir")
				os.MkdirAll(subdir, 0755)
				os.WriteFile(filepath.Join(dir, "file1.go"), []byte("package test"), 0644)
				os.WriteFile(filepath.Join(subdir, "file2.go"), []byte("package test"), 0644)
				return nil
			},
			expectedCount: 2,
		},
		{
			name: "mixed file types",
			setup: func(dir string) error {
				os.WriteFile(filepath.Join(dir, "file.go"), []byte("package test"), 0644)
				os.WriteFile(filepath.Join(dir, "file.txt"), []byte("text"), 0644)
				os.WriteFile(filepath.Join(dir, "file.json"), []byte("{}"), 0644)
				return nil
			},
			expectedCount: 1,
		},
		{
			name: "no go files",
			setup: func(dir string) error {
				os.WriteFile(filepath.Join(dir, "file.txt"), []byte("text"), 0644)
				return nil
			},
			expectedCount: 0,
		},
		{
			name: "empty directory",
			setup: func(dir string) error {
				return nil
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			processor := NewFormatterProcessor(false)
			files, err := processor.findGoFiles(tmpDir)

			if err != nil {
				t.Errorf("findGoFiles() error = %v", err)
				return
			}

			if len(files) != tt.expectedCount {
				t.Errorf("findGoFiles() found %d files, want %d", len(files), tt.expectedCount)
			}

			// Verify all found files are .go files
			for _, file := range files {
				if filepath.Ext(file) != ".go" {
					t.Errorf("findGoFiles() returned non-Go file: %s", file)
				}
			}
		})
	}
}

func TestFormatterProcessorProcess(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(string) ProcessSpec
		wantErr bool
	}{
		{
			name: "format single file",
			setup: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				// Create a Go file with poor formatting
				goFile := filepath.Join(clientPath, "test.go")
				content := "package test\n\nfunc  Test()   {}\n"
				os.WriteFile(goFile, []byte(content), 0644)

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    "/tmp/spec.json",
					PackageName: "testpkg",
				}
			},
			wantErr: false,
		},
		{
			name: "format multiple files",
			setup: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				// Create multiple Go files
				for i := 1; i <= 3; i++ {
					goFile := filepath.Join(clientPath, filepath.Join("test", string(rune('0'+i))+".go"))
					content := "package test\n\nfunc Test()   {}\n"
					os.WriteFile(goFile, []byte(content), 0644)
				}

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    "/tmp/spec.json",
					PackageName: "testpkg",
				}
			},
			wantErr: false,
		},
		{
			name: "no go files (should not error)",
			setup: func(tmpDir string) ProcessSpec {
				clientPath := filepath.Join(tmpDir, "client")
				os.MkdirAll(clientPath, 0755)

				return ProcessSpec{
					ClientPath:  clientPath,
					ServiceName: "testservice",
					SpecPath:    "/tmp/spec.json",
					PackageName: "testpkg",
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			spec := tt.setup(tmpDir)

			processor := NewFormatterProcessor(false)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := processor.Process(ctx, spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatterProcessorProcessWithSimplify(t *testing.T) {
	tmpDir := t.TempDir()
	clientPath := filepath.Join(tmpDir, "client")
	os.MkdirAll(clientPath, 0755)

	// Create a Go file
	goFile := filepath.Join(clientPath, "test.go")
	content := "package test\n\nfunc Test() {}\n"
	os.WriteFile(goFile, []byte(content), 0644)

	spec := ProcessSpec{
		ClientPath:  clientPath,
		ServiceName: "testservice",
		SpecPath:    "/tmp/spec.json",
		PackageName: "testpkg",
	}

	// Test with simplify enabled
	processor := NewFormatterProcessor(true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := processor.Process(ctx, spec)
	if err != nil {
		t.Errorf("Process() with simplify error = %v", err)
	}
}

func TestFormatterProcessorImplementsInterface(t *testing.T) {
	// Verify FormatterProcessor implements PostProcessor interface
	var _ PostProcessor = (*FormatterProcessor)(nil)
}
