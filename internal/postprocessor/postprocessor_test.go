package postprocessor

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockPostProcessor is a mock implementation for testing
type MockPostProcessor struct {
	name        string
	shouldError bool
	processed   bool
}

func NewMockPostProcessor(name string, shouldError bool) *MockPostProcessor {
	return &MockPostProcessor{
		name:        name,
		shouldError: shouldError,
		processed:   false,
	}
}

func (m *MockPostProcessor) Name() string {
	return m.name
}

func (m *MockPostProcessor) Process(ctx context.Context, spec ProcessSpec) error {
	m.processed = true
	if m.shouldError {
		return fmt.Errorf("mock error from %s", m.name)
	}
	return nil
}

func TestNewChain(t *testing.T) {
	chain := NewChain()

	if chain == nil {
		t.Fatal("NewChain() returned nil")
	}

	if chain.Count() != 0 {
		t.Errorf("NewChain() count = %d, want 0", chain.Count())
	}
}

func TestChainAdd(t *testing.T) {
	tests := []struct {
		name      string
		processor PostProcessor
		wantErr   bool
	}{
		{
			name:      "add valid processor",
			processor: NewMockPostProcessor("test", false),
			wantErr:   false,
		},
		{
			name:      "add nil processor",
			processor: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewChain()
			err := chain.Add(tt.processor)

			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.processor != nil {
				if chain.Count() != 1 {
					t.Errorf("Count() = %d after Add(), want 1", chain.Count())
				}
			}
		})
	}
}

func TestChainAddMultiple(t *testing.T) {
	chain := NewChain()

	processors := []PostProcessor{
		NewMockPostProcessor("first", false),
		NewMockPostProcessor("second", false),
		NewMockPostProcessor("third", false),
	}

	for _, p := range processors {
		if err := chain.Add(p); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	if chain.Count() != 3 {
		t.Errorf("Count() = %d, want 3", chain.Count())
	}
}

func TestChainProcess(t *testing.T) {
	tests := []struct {
		name       string
		processors []PostProcessor
		wantErr    bool
		errMsg     string
	}{
		{
			name: "empty chain",
			processors: []PostProcessor{},
			wantErr:    false,
		},
		{
			name: "single successful processor",
			processors: []PostProcessor{
				NewMockPostProcessor("test", false),
			},
			wantErr: false,
		},
		{
			name: "multiple successful processors",
			processors: []PostProcessor{
				NewMockPostProcessor("first", false),
				NewMockPostProcessor("second", false),
				NewMockPostProcessor("third", false),
			},
			wantErr: false,
		},
		{
			name: "processor fails",
			processors: []PostProcessor{
				NewMockPostProcessor("success", false),
				NewMockPostProcessor("failing", true),
				NewMockPostProcessor("never-runs", false),
			},
			wantErr: true,
			errMsg:  "failing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewChain()
			for _, p := range tt.processors {
				chain.Add(p)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			spec := ProcessSpec{
				ClientPath:  "/tmp/client",
				ServiceName: "testservice",
				SpecPath:    "/tmp/spec.json",
				PackageName: "testpkg",
			}

			err := chain.Process(ctx, spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Process() error = %q, should contain %q", err.Error(), tt.errMsg)
				}
			}

			// Verify processors were called in order until failure
			for i, p := range tt.processors {
				mock, ok := p.(*MockPostProcessor)
				if !ok {
					continue
				}

				if tt.wantErr && mock.shouldError {
					// This processor should have been called (it's the one that failed)
					if !mock.processed {
						t.Errorf("Processor %d (%s) should have been called but wasn't", i, mock.name)
					}
					// Processors after this shouldn't be called
					break
				} else if tt.wantErr {
					// Processors before the failing one should be called
					if !mock.processed {
						t.Errorf("Processor %d (%s) should have been called before failure", i, mock.name)
					}
				} else {
					// All processors should be called
					if !mock.processed {
						t.Errorf("Processor %d (%s) should have been called", i, mock.name)
					}
				}
			}
		})
	}
}

func TestChainProcessCancellation(t *testing.T) {
	chain := NewChain()
	chain.Add(NewMockPostProcessor("test", false))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	spec := ProcessSpec{
		ClientPath:  "/tmp/client",
		ServiceName: "testservice",
		SpecPath:    "/tmp/spec.json",
		PackageName: "testpkg",
	}

	err := chain.Process(ctx, spec)

	if err == nil {
		t.Error("Process() should fail with cancelled context")
	}

	if !contains(err.Error(), "cancelled") {
		t.Errorf("Error should mention cancellation, got: %v", err)
	}
}

func TestChainCount(t *testing.T) {
	chain := NewChain()

	if chain.Count() != 0 {
		t.Errorf("Count() on empty chain = %d, want 0", chain.Count())
	}

	chain.Add(NewMockPostProcessor("first", false))
	if chain.Count() != 1 {
		t.Errorf("Count() after one Add() = %d, want 1", chain.Count())
	}

	chain.Add(NewMockPostProcessor("second", false))
	if chain.Count() != 2 {
		t.Errorf("Count() after two Add() = %d, want 2", chain.Count())
	}
}

func TestChainClear(t *testing.T) {
	chain := NewChain()
	chain.Add(NewMockPostProcessor("first", false))
	chain.Add(NewMockPostProcessor("second", false))

	if chain.Count() != 2 {
		t.Fatalf("Count() before Clear() = %d, want 2", chain.Count())
	}

	chain.Clear()

	if chain.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", chain.Count())
	}
}

func TestChainList(t *testing.T) {
	chain := NewChain()

	// Empty chain
	list := chain.List()
	if len(list) != 0 {
		t.Errorf("List() on empty chain = %v, want empty slice", list)
	}

	// Add processors
	chain.Add(NewMockPostProcessor("first", false))
	chain.Add(NewMockPostProcessor("second", false))
	chain.Add(NewMockPostProcessor("third", false))

	list = chain.List()
	if len(list) != 3 {
		t.Errorf("List() length = %d, want 3", len(list))
	}

	expected := []string{"first", "second", "third"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("List()[%d] = %q, want %q", i, list[i], name)
		}
	}
}

func TestProcessSpec(t *testing.T) {
	spec := ProcessSpec{
		ClientPath:  "/path/to/client",
		ServiceName: "testservice",
		SpecPath:    "/path/to/spec.json",
		PackageName: "testpkg",
	}

	if spec.ClientPath != "/path/to/client" {
		t.Errorf("ClientPath = %q", spec.ClientPath)
	}

	if spec.ServiceName != "testservice" {
		t.Errorf("ServiceName = %q", spec.ServiceName)
	}

	if spec.SpecPath != "/path/to/spec.json" {
		t.Errorf("SpecPath = %q", spec.SpecPath)
	}

	if spec.PackageName != "testpkg" {
		t.Errorf("PackageName = %q", spec.PackageName)
	}
}

// Helper function
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
