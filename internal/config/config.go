package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
)

// Config holds all configuration parameters for the application
type Config struct {
	// SpecsDir is the directory containing OpenAPI specification files
	SpecsDir string `mapstructure:"specs_dir"`

	// OutputDir is the base directory where generated clients will be stored
	OutputDir string `mapstructure:"output_dir"`

	// TargetServices is a regular expression pattern to filter services
	// Empty string matches all services
	TargetServices string `mapstructure:"target_services"`

	// ContinueOnError allows generation to continue even if some specs fail
	// Default: false (fail fast on first error)
	ContinueOnError bool `mapstructure:"continue_on_error"`
}

// LoadConfig initializes Viper and loads configuration from application.yml
// with the ability to override via environment variables
func LoadConfig() (Config, error) {
	v := viper.New()

	// Set up config file support with absolute paths
	resourcesDir := paths.GetResourcesDir()

	v.SetConfigName("application")
	v.SetConfigType("yml")
	v.AddConfigPath(resourcesDir)

	// Also check user home directory
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".openapi-go"))
	}

	// Enable automatic environment variable binding
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	log.Printf("Using config file: %s", v.ConfigFileUsed())

	// Unmarshal config into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Convert relative paths to absolute paths
	cfg.SpecsDir = paths.MakeAbsolutePath(cfg.SpecsDir)
	cfg.OutputDir = paths.MakeAbsolutePath(cfg.OutputDir)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (cfg *Config) Validate() error {
	// Validate SpecsDir exists
	if cfg.SpecsDir == "" {
		return fmt.Errorf("specs_dir is required")
	}
	if err := paths.EnsurePathExists(cfg.SpecsDir); err != nil {
		return fmt.Errorf("specs_dir validation failed: %w", err)
	}

	// Validate OutputDir
	if cfg.OutputDir == "" {
		return fmt.Errorf("output_dir is required")
	}

	// Create output directory if it doesn't exist and check if writable
	if err := paths.EnsureDirectoryWritable(cfg.OutputDir); err != nil {
		return fmt.Errorf("output_dir validation failed: %w", err)
	}

	// Validate TargetServices regex
	if cfg.TargetServices != "" {
		if _, err := regexp.Compile(cfg.TargetServices); err != nil {
			return fmt.Errorf("target_services is not a valid regex: %w", err)
		}
	}

	return nil
}

// LogConfiguration logs the current configuration parameters
func LogConfiguration(cfg Config) {
	log.Printf("Configuration loaded:")
	log.Printf("  Repository root: %s", paths.GetRepositoryRoot())
	log.Printf("  Specs directory: %s", cfg.SpecsDir)
	log.Printf("  Output directory: %s", cfg.OutputDir)
	log.Printf("  Target services: %s", cfg.TargetServices)
	log.Printf("  Continue on error: %v", cfg.ContinueOnError)
	log.Printf("  Ogen config: %s", paths.GetOgenConfigPath())
}
