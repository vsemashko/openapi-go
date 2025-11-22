package generator

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.Count() != 0 {
		t.Errorf("NewRegistry() count = %d, want 0", registry.Count())
	}

	if registry.defaultGenerator != "" {
		t.Errorf("NewRegistry() defaultGenerator = %q, want empty", registry.defaultGenerator)
	}
}

func TestRegistryRegister(t *testing.T) {
	tests := []struct {
		name    string
		gen     Generator
		wantErr bool
		errMsg  string
	}{
		{
			name:    "register valid generator",
			gen:     NewOgenGenerator(),
			wantErr: false,
		},
		{
			name:    "register nil generator",
			gen:     nil,
			wantErr: true,
			errMsg:  "cannot register nil generator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			err := registry.Register(tt.gen)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Register() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}

			// If registration succeeded, verify it's in the registry
			if err == nil && tt.gen != nil {
				if registry.Count() != 1 {
					t.Errorf("Registry count = %d after registration, want 1", registry.Count())
				}

				// Should be set as default (first generator)
				if registry.defaultGenerator != tt.gen.Name() {
					t.Errorf("Default generator = %q, want %q", registry.defaultGenerator, tt.gen.Name())
				}
			}
		})
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	registry := NewRegistry()
	gen := NewOgenGenerator()

	// First registration should succeed
	err := registry.Register(gen)
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	// Second registration of same generator should fail
	err = registry.Register(gen)
	if err == nil {
		t.Error("Register() should fail for duplicate generator")
	}

	if !contains(err.Error(), "already registered") {
		t.Errorf("Error should mention 'already registered', got: %v", err)
	}
}

func TestRegistryGet(t *testing.T) {
	registry := NewRegistry()
	gen := NewOgenGenerator()
	registry.Register(gen)

	tests := []struct {
		name       string
		genName    string
		wantErr    bool
		wantNil    bool
		errContains string
	}{
		{
			name:    "get existing generator",
			genName: "ogen",
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "get non-existent generator",
			genName:     "nonexistent",
			wantErr:     true,
			wantNil:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.Get(tt.genName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (result == nil) != tt.wantNil {
				t.Errorf("Get() result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Get() error = %q, should contain %q", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestRegistryGetDefault(t *testing.T) {
	t.Run("get default when none registered", func(t *testing.T) {
		registry := NewRegistry()
		gen, err := registry.GetDefault()

		if err == nil {
			t.Error("GetDefault() should fail when no generators registered")
		}

		if gen != nil {
			t.Error("GetDefault() should return nil when no generators registered")
		}
	})

	t.Run("get default after registration", func(t *testing.T) {
		registry := NewRegistry()
		ogen := NewOgenGenerator()
		registry.Register(ogen)

		gen, err := registry.GetDefault()
		if err != nil {
			t.Fatalf("GetDefault() error = %v", err)
		}

		if gen == nil {
			t.Fatal("GetDefault() returned nil")
		}

		if gen.Name() != "ogen" {
			t.Errorf("GetDefault() name = %q, want %q", gen.Name(), "ogen")
		}
	})
}

func TestRegistrySetDefault(t *testing.T) {
	registry := NewRegistry()
	ogen := NewOgenGenerator()
	registry.Register(ogen)

	tests := []struct {
		name    string
		genName string
		wantErr bool
	}{
		{
			name:    "set to existing generator",
			genName: "ogen",
			wantErr: false,
		},
		{
			name:    "set to non-existent generator",
			genName: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.SetDefault(tt.genName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetDefault() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if registry.defaultGenerator != tt.genName {
					t.Errorf("defaultGenerator = %q, want %q", registry.defaultGenerator, tt.genName)
				}
			}
		})
	}
}

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("List() on empty registry = %v, want empty slice", list)
	}

	// Add generator
	ogen := NewOgenGenerator()
	registry.Register(ogen)

	list = registry.List()
	if len(list) != 1 {
		t.Errorf("List() length = %d, want 1", len(list))
	}

	if list[0] != "ogen" {
		t.Errorf("List()[0] = %q, want %q", list[0], "ogen")
	}
}

func TestRegistryCount(t *testing.T) {
	registry := NewRegistry()

	if registry.Count() != 0 {
		t.Errorf("Count() on empty registry = %d, want 0", registry.Count())
	}

	// Add generators
	ogen := NewOgenGenerator()
	registry.Register(ogen)

	if registry.Count() != 1 {
		t.Errorf("Count() after one registration = %d, want 1", registry.Count())
	}
}

func TestRegistryClear(t *testing.T) {
	registry := NewRegistry()
	ogen := NewOgenGenerator()
	registry.Register(ogen)

	// Verify registry has content
	if registry.Count() == 0 {
		t.Fatal("Registry should have generators before Clear()")
	}

	// Clear
	registry.Clear()

	// Verify empty
	if registry.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", registry.Count())
	}

	if registry.defaultGenerator != "" {
		t.Errorf("defaultGenerator after Clear() = %q, want empty", registry.defaultGenerator)
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
