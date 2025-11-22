package config

import (
	"log"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
)

// LogConfiguration logs the current configuration parameters using structured logging
// If logger is provided, uses structured logging; otherwise uses standard log package
func LogConfiguration(cfg Config, optionalLogger ...interface{}) {
	var logger interface{}
	if len(optionalLogger) > 0 {
		logger = optionalLogger[0]
	}

	// Support both structured logger and nil (fallback to log package)
	type structuredLogger interface {
		Info(msg string, args ...any)
	}

	if slog, ok := logger.(structuredLogger); ok {
		// Use structured logging
		slog.Info("Configuration loaded",
			"repository_root", paths.GetRepositoryRoot(),
			"specs_directory", cfg.SpecsDir,
			"output_directory", cfg.OutputDir,
			"target_services", cfg.TargetServices,
			"continue_on_error", cfg.ContinueOnError,
			"worker_count", cfg.WorkerCount,
			"enable_cache", cfg.EnableCache,
			"cache_directory", cfg.CacheDir,
			"spec_file_patterns", cfg.SpecFilePatterns,
			"log_level", cfg.LogLevel,
			"log_format", cfg.LogFormat,
			"ogen_config", paths.GetOgenConfigPath(),
		)
	} else {
		// Fallback to standard logging (backward compatibility)
		log.Printf("Configuration loaded:")
		log.Printf("  Repository root: %s", paths.GetRepositoryRoot())
		log.Printf("  Specs directory: %s", cfg.SpecsDir)
		log.Printf("  Output directory: %s", cfg.OutputDir)
		log.Printf("  Target services: %s", cfg.TargetServices)
		log.Printf("  Continue on error: %v", cfg.ContinueOnError)
		log.Printf("  Worker count: %d", cfg.WorkerCount)
		log.Printf("  Enable cache: %v", cfg.EnableCache)
		log.Printf("  Cache directory: %s", cfg.CacheDir)
		log.Printf("  Spec file patterns: %v", cfg.SpecFilePatterns)
		log.Printf("  Log level: %s", cfg.LogLevel)
		log.Printf("  Log format: %s", cfg.LogFormat)
		log.Printf("  Ogen config: %s", paths.GetOgenConfigPath())
	}
}
