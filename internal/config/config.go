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

	// WorkerCount is the number of parallel workers for spec processing
	// Default: 4
	WorkerCount int `mapstructure:"worker_count"`

	// EnableCache enables caching of generated clients to skip regeneration
	// Default: true
	EnableCache bool `mapstructure:"enable_cache"`

	// CacheDir is the directory where cache metadata is stored
	// Default: .openapi-cache
	CacheDir string `mapstructure:"cache_dir"`

	// SpecFilePatterns are the filenames to look for when discovering OpenAPI specs
	// Default: ["openapi.json", "openapi.yaml", "openapi.yml"]
	SpecFilePatterns []string `mapstructure:"spec_file_patterns"`

	// LogLevel sets the logging level (debug, info, warn, error)
	// Default: info
	LogLevel string `mapstructure:"log_level"`

	// LogFormat sets the log output format (json, text)
	// Default: json
	LogFormat string `mapstructure:"log_format"`
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

	// Set defaults for optional fields
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}

	// Set EnableCache default to true (caching enabled by default)
	// Note: Viper unmarshals false as zero value, so we need explicit handling
	// If not set in config, enable cache by default
	v.SetDefault("enable_cache", true)
	cfg.EnableCache = v.GetBool("enable_cache")

	if cfg.CacheDir == "" {
		cfg.CacheDir = ".openapi-cache"
	}

	// Set default spec file patterns if not specified
	if len(cfg.SpecFilePatterns) == 0 {
		cfg.SpecFilePatterns = []string{"openapi.json", "openapi.yaml", "openapi.yml"}
	}

	// Set default log level and format
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = "json"
	}

	// Convert relative paths to absolute paths
	cfg.SpecsDir = paths.MakeAbsolutePath(cfg.SpecsDir)
	cfg.OutputDir = paths.MakeAbsolutePath(cfg.OutputDir)
	cfg.CacheDir = paths.MakeAbsolutePath(cfg.CacheDir)

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

// LogConfiguration is now in config_logging.go to support structured logging
