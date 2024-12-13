package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
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
}

// LoadConfig initializes Viper and loads configuration from application.yml
// with the ability to override via environment variables
func LoadConfig() (Config, error) {
	v := viper.New()

	// Set up config file support
	v.SetConfigName("application")
	v.SetConfigType("yml")
	v.AddConfigPath("./resources")
	v.AddConfigPath("$HOME/.openapi-go")

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

	return cfg, nil
}

// LogConfiguration logs the current configuration parameters
func LogConfiguration(cfg Config) {
	log.Printf("Processing OpenAPI specs with configuration:")
	log.Printf("- Specs directory: %s", cfg.SpecsDir)
	log.Printf("- Output directory: %s", cfg.OutputDir)
	log.Printf("- Target services: %s", cfg.TargetServices)
}
