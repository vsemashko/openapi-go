package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/cache"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/config"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/generator"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/metrics"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/paths"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/spec"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/validator"
	"gitlab.stashaway.com/vladimir.semashko/openapi-go/internal/worker"
)

var (
	// defaultGenerator is the generator used for code generation
	// Can be overridden for testing or to support different generators
	defaultGenerator generator.Generator = generator.NewOgenGenerator()
)

// ProcessingResult contains the results of processing OpenAPI specs
type ProcessingResult struct {
	TotalSpecs   int
	SuccessCount int
	FailedSpecs  []SpecFailure
}

// SpecFailure represents a failed spec generation
type SpecFailure struct {
	SpecPath    string
	ServiceName string
	Error       error
}

// ProcessOpenAPISpecs processes OpenAPI specifications and generates client code.
// It searches for OpenAPI specs in the specified directory that match the targetServices pattern,
// then generates Go client code for each spec using the configured generator.
//
// Parameters:
// - ctx: Context for cancellation and timeouts
// - cfg: Configuration containing specs directory, output directory, and target services pattern
// - optionalLogger: Optional structured logger (if not provided, uses standard log package)
//
// Returns an error if the process fails at any stage.
func ProcessOpenAPISpecs(ctx context.Context, cfg config.Config, optionalLogger ...interface{}) error {
	// Extract logger if provided (for future migration to structured logging)
	// For now, we still use log.Printf in most places, but this allows gradual migration
	var _ interface{} = nil
	if len(optionalLogger) > 0 {
		_ = optionalLogger[0]
		// Future: Use structured logger throughout
	}

	// Initialize metrics collector
	metricsCollector := metrics.NewCollector()
	defer func() {
		// Finalize and export metrics
		metricsCollector.Finalize()

		// Export to file
		metricsPath := filepath.Join(cfg.OutputDir, ".openapi-metrics.json")
		if err := metricsCollector.Export(metricsPath); err != nil {
			log.Printf("Warning: Failed to export metrics: %v", err)
		} else {
			log.Printf("Metrics exported to: %s", metricsPath)
		}

		// Log summary
		log.Printf("%s", metricsCollector.Summary())
		log.Printf("Success rate: %.1f%%", metricsCollector.SuccessRate())
		log.Printf("Cache hit rate: %.1f%%", metricsCollector.CacheHitRate())
	}()

	// Setup the client output directory
	clientOutputDir := filepath.Join(cfg.OutputDir, "clients")
	if err := os.MkdirAll(clientOutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create client output directory: %w", err)
	}

	// Find OpenAPI specs
	specs, err := findOpenAPISpecs(cfg.SpecsDir, cfg.TargetServices, cfg.SpecFilePatterns)
	if err != nil {
		return err
	}

	// Validate specs if validation is enabled
	if cfg.Validator.Enabled {
		log.Printf("Validating %d OpenAPI specs...", len(specs))
		if err := validateSpecs(specs, cfg.Validator, cfg.ContinueOnError); err != nil {
			return fmt.Errorf("spec validation failed: %w", err)
		}
		log.Printf("All specs validated successfully")
	}

	// Initialize cache if enabled
	var specCache *cache.Cache
	if cfg.EnableCache {
		specCache, err = cache.NewCache(cache.Config{CacheDir: cfg.CacheDir})
		if err != nil {
			log.Printf("Warning: Failed to initialize cache, proceeding without caching: %v", err)
			specCache = nil
		} else {
			// Prune invalid cache entries
			pruned, err := specCache.PruneInvalid()
			if err != nil {
				log.Printf("Warning: Failed to prune cache: %v", err)
			} else if pruned > 0 {
				log.Printf("Pruned %d invalid cache entries", pruned)
			}
		}
	}

	// Generate clients in parallel
	result, err := generateClients(ctx, specs, cfg.OutputDir, cfg.ContinueOnError, cfg.WorkerCount, specCache, metricsCollector)
	if err != nil {
		return err
	}

	// Log results
	logProcessingResult(result)

	// Return error if any specs failed (unless continue-on-error is enabled)
	if !cfg.ContinueOnError && result.SuccessCount < result.TotalSpecs {
		return fmt.Errorf("failed to generate %d/%d clients",
			len(result.FailedSpecs), result.TotalSpecs)
	}

	return nil
}

// findOpenAPISpecs searches for OpenAPI specs in the given directory.
func findOpenAPISpecs(specsDir string, targetServices string, specFilePatterns []string) ([]string, error) {
	// Compile service regex for filtering
	serviceRegex, err := compileServiceRegex(targetServices)
	if err != nil {
		return nil, err
	}

	// If no patterns specified, use default
	if len(specFilePatterns) == 0 {
		specFilePatterns = []string{"openapi.json", "openapi.yaml", "openapi.yml"}
	}

	var specs []string

	err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		// Skip directories and errors
		if err != nil || info.IsDir() {
			return nil
		}

		// Check if filename matches any of the spec file patterns
		filename := filepath.Base(path)
		isSpecFile := false
		for _, pattern := range specFilePatterns {
			if filename == pattern {
				isSpecFile = true
				break
			}
		}

		if !isSpecFile {
			return nil
		}

		// Check if service name matches the filter
		serviceDir := filepath.Base(filepath.Dir(path))
		if !serviceRegex.MatchString(serviceDir) {
			return nil
		}

		specs = append(specs, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find OpenAPI specs: %w", err)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("no OpenAPI specs found for target services")
	}

	log.Printf("Found %d OpenAPI specs matching the criteria", len(specs))
	return specs, nil
}

// generateClients generates clients for all found OpenAPI specs using parallel processing.
func generateClients(ctx context.Context, specs []string, outputDir string, continueOnError bool, workerCount int, specCache *cache.Cache, metricsCollector *metrics.Collector) (*ProcessingResult, error) {
	result := &ProcessingResult{
		TotalSpecs:   len(specs),
		SuccessCount: 0,
		FailedSpecs:  []SpecFailure{},
	}

	// If only one spec or worker count is 1, process sequentially
	if len(specs) == 1 || workerCount == 1 {
		return generateClientsSequential(ctx, specs, outputDir, continueOnError, specCache, metricsCollector)
	}

	log.Printf("Processing %d specs with %d parallel workers", len(specs), workerCount)

	// Create worker pool
	pool := worker.NewPool(worker.Config{
		WorkerCount:   workerCount,
		TaskQueueSize: len(specs),
	})

	// Create tasks for each spec
	tasks := make([]worker.Task, 0, len(specs))
	for _, specPath := range specs {
		// Capture variables for closure
		currentSpecPath := specPath
		serviceDir := filepath.Base(filepath.Dir(currentSpecPath))
		serviceName := normalizeServiceName(serviceDir)
		folderName := serviceName + "sdk"

		task := worker.Task{
			ID: serviceName,
			Execute: func(taskCtx context.Context) error {
				// Start timing for metrics
				startTime := time.Now()

				// Parse spec and create operation fingerprint (for both caching and validation)
				parsedSpec, parseErr := spec.ParseSpecFile(currentSpecPath)
				var fingerprint *spec.SpecFingerprint
				if parseErr == nil {
					fingerprint, _ = spec.CreateSpecFingerprint(currentSpecPath, parsedSpec)
				}

				// Check cache if available (using incremental validation)
				if specCache != nil {
					// Check cache using incremental validation
					valid, comparison, err := specCache.IsValidIncremental(currentSpecPath, defaultGenerator.Version(), fingerprint)
					if err != nil {
						log.Printf("Warning: Cache check failed for %s: %v", serviceName, err)
					} else if valid {
						log.Printf("⚡ Using cached client for %s (no operation changes detected)", folderName)

						// Record cached metric
						metricsCollector.RecordSpec(metrics.SpecMetric{
							SpecPath:    currentSpecPath,
							ServiceName: serviceName,
							Success:     true,
							Cached:      true,
							DurationMs:  time.Since(startTime).Milliseconds(),
							GeneratedAt: time.Now(),
						})
						return nil
					} else if comparison != nil && comparison.HasChanges() {
						// Log what changed
						log.Printf("Regenerating %s: %s", serviceName, comparison.Summary())
					}
				}

				log.Printf("Processing service: %s (spec: %s)", serviceName, currentSpecPath)
				clientPath := filepath.Join(outputDir, "clients", folderName)

				// Generate client
				genErr := generateClientForSpec(taskCtx, currentSpecPath, serviceName, folderName, outputDir)
				duration := time.Since(startTime).Milliseconds()

				if genErr != nil {
					// Record failed metric
					metricsCollector.RecordSpec(metrics.SpecMetric{
						SpecPath:    currentSpecPath,
						ServiceName: serviceName,
						Success:     false,
						Cached:      false,
						DurationMs:  duration,
						Error:       genErr.Error(),
						GeneratedAt: time.Now(),
					})
					return genErr
				}

				// Record successful metric
				metricsCollector.RecordSpec(metrics.SpecMetric{
					SpecPath:    currentSpecPath,
					ServiceName: serviceName,
					Success:     true,
					Cached:      false,
					DurationMs:  duration,
					GeneratedAt: time.Now(),
				})

				// Update cache on success with operation fingerprint
				if specCache != nil {
					if err := specCache.SetWithFingerprint(currentSpecPath, clientPath, serviceName, defaultGenerator.Version(), fingerprint); err != nil {
						log.Printf("Warning: Failed to update cache for %s: %v", serviceName, err)
					}
				}

				return nil
			},
		}
		tasks = append(tasks, task)
	}

	// Process all tasks in parallel
	results, err := pool.ProcessBatch(ctx, tasks)
	if err != nil {
		return result, fmt.Errorf("parallel processing failed: %w", err)
	}

	// Collect results with thread-safe access
	var mu sync.Mutex
	for _, taskResult := range results {
		if taskResult.Error != nil {
			// Find the corresponding spec path
			var specPath string
			for _, spec := range specs {
				serviceDir := filepath.Base(filepath.Dir(spec))
				serviceName := normalizeServiceName(serviceDir)
				if serviceName == taskResult.TaskID {
					specPath = spec
					break
				}
			}

			failure := SpecFailure{
				SpecPath:    specPath,
				ServiceName: taskResult.TaskID,
				Error:       taskResult.Error,
			}

			mu.Lock()
			result.FailedSpecs = append(result.FailedSpecs, failure)
			mu.Unlock()

			log.Printf("❌ Failed to generate client for %ssdk: %v", taskResult.TaskID, taskResult.Error)

			// Fail fast unless continue-on-error is enabled
			if !continueOnError {
				return result, fmt.Errorf("generation failed for %s: %w", taskResult.TaskID, taskResult.Error)
			}
		} else {
			mu.Lock()
			result.SuccessCount++
			mu.Unlock()
			log.Printf("✅ Successfully generated client for %ssdk", taskResult.TaskID)
		}
	}

	return result, nil
}

// generateClientsSequential generates clients sequentially (fallback for single spec or single worker).
func generateClientsSequential(ctx context.Context, specs []string, outputDir string, continueOnError bool, specCache *cache.Cache, metricsCollector *metrics.Collector) (*ProcessingResult, error) {
	result := &ProcessingResult{
		TotalSpecs:   len(specs),
		SuccessCount: 0,
		FailedSpecs:  []SpecFailure{},
	}

	for _, specPath := range specs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("generation cancelled: %w", ctx.Err())
		default:
		}

		serviceDir := filepath.Base(filepath.Dir(specPath))
		serviceName := normalizeServiceName(serviceDir)
		folderName := serviceName + "sdk"
		clientPath := filepath.Join(outputDir, "clients", folderName)

		// Start timing for metrics
		startTime := time.Now()

		// Parse spec and create operation fingerprint (for both caching and validation)
		parsedSpec, parseErr := spec.ParseSpecFile(specPath)
		var fingerprint *spec.SpecFingerprint
		if parseErr == nil {
			fingerprint, _ = spec.CreateSpecFingerprint(specPath, parsedSpec)
		}

		// Check cache if available (using incremental validation)
		if specCache != nil {
			valid, comparison, err := specCache.IsValidIncremental(specPath, defaultGenerator.Version(), fingerprint)
			if err != nil {
				log.Printf("Warning: Cache check failed for %s: %v", serviceName, err)
			} else if valid {
				log.Printf("⚡ Using cached client for %s (no operation changes detected)", folderName)
				result.SuccessCount++

				// Record cached metric
				metricsCollector.RecordSpec(metrics.SpecMetric{
					SpecPath:    specPath,
					ServiceName: serviceName,
					Success:     true,
					Cached:      true,
					DurationMs:  time.Since(startTime).Milliseconds(),
					GeneratedAt: time.Now(),
				})
				continue
			} else if comparison != nil && comparison.HasChanges() {
				// Log what changed
				log.Printf("Regenerating %s: %s", serviceName, comparison.Summary())
			}
		}

		log.Printf("Processing service: %s (spec: %s)", serviceName, specPath)

		err := generateClientForSpec(ctx, specPath, serviceName, folderName, outputDir)
		duration := time.Since(startTime).Milliseconds()

		if err != nil {
			failure := SpecFailure{
				SpecPath:    specPath,
				ServiceName: serviceName,
				Error:       err,
			}
			result.FailedSpecs = append(result.FailedSpecs, failure)

			log.Printf("❌ Failed to generate client for %s: %v", folderName, err)

			// Record failed metric
			metricsCollector.RecordSpec(metrics.SpecMetric{
				SpecPath:    specPath,
				ServiceName: serviceName,
				Success:     false,
				Cached:      false,
				DurationMs:  duration,
				Error:       err.Error(),
				GeneratedAt: time.Now(),
			})

			// Fail fast unless continue-on-error is enabled
			if !continueOnError {
				return result, fmt.Errorf("generation failed for %s: %w", serviceName, err)
			}
		} else {
			result.SuccessCount++
			log.Printf("✅ Successfully generated client for %s", folderName)

			// Record successful metric
			metricsCollector.RecordSpec(metrics.SpecMetric{
				SpecPath:    specPath,
				ServiceName: serviceName,
				Success:     true,
				Cached:      false,
				DurationMs:  duration,
				GeneratedAt: time.Now(),
			})

			// Update cache on success with operation fingerprint
			if specCache != nil {
				if err := specCache.SetWithFingerprint(specPath, clientPath, serviceName, defaultGenerator.Version(), fingerprint); err != nil {
					log.Printf("Warning: Failed to update cache for %s: %v", serviceName, err)
				}
			}
		}
	}

	return result, nil
}

// logProcessingResult logs a summary of the processing results
func logProcessingResult(result *ProcessingResult) {
	log.Printf("=====================================")
	log.Printf("SDK Generation Summary")
	log.Printf("=====================================")
	log.Printf("Total specs:    %d", result.TotalSpecs)
	log.Printf("Successful:     %d", result.SuccessCount)
	log.Printf("Failed:         %d", len(result.FailedSpecs))

	if len(result.FailedSpecs) > 0 {
		log.Printf("-------------------------------------")
		log.Printf("Failed specs:")
		for _, failure := range result.FailedSpecs {
			log.Printf("  - %s: %v", failure.ServiceName, failure.Error)
		}
	}
	log.Printf("=====================================")
}

// generateClientForSpec generates a client for a single OpenAPI spec.
func generateClientForSpec(ctx context.Context, specPath, serviceName, folderName, outputDir string) error {
	// Create the client directory
	clientPath := filepath.Join(outputDir, "clients", folderName)
	if err := os.MkdirAll(clientPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create client directory for %s: %w", serviceName, err)
	}

	// Clean existing files in the client directory
	log.Printf("Cleaning existing files for %s...", folderName)
	if err := cleanDirectory(clientPath); err != nil {
		return fmt.Errorf("failed to clean client directory for %s: %w", serviceName, err)
	}

	// Run the client generator
	if err := runGenerator(ctx, folderName, specPath, clientPath); err != nil {
		return err
	}

	// Apply post-processors to the generated client
	log.Printf("Applying post-processors for %s...", folderName)
	if err := ApplyPostProcessors(ctx, clientPath, folderName, specPath); err != nil {
		return fmt.Errorf("failed to apply post-processors for %s: %w", folderName, err)
	}

	log.Printf("Successfully generated client for %s", folderName)
	return nil
}

// runGenerator executes the configured generator to create client code from an OpenAPI spec.
func runGenerator(ctx context.Context, serviceName, specPath, outputDir string) error {
	log.Printf("Generating client for %s using %s...", serviceName, defaultGenerator.Name())

	// Create generate spec
	spec := generator.GenerateSpec{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: serviceName,
		ConfigPath:  paths.GetOgenConfigPath(),
		Clean:       true,
	}

	// Generate client code
	if err := defaultGenerator.Generate(ctx, spec); err != nil {
		return fmt.Errorf("generation failed for %s: %w", serviceName, err)
	}

	return nil
}

// SetGenerator allows overriding the default generator (useful for testing)
func SetGenerator(gen generator.Generator) {
	if gen != nil {
		defaultGenerator = gen
	}
}

// validateSpecs validates all OpenAPI specs before generation
func validateSpecs(specs []string, validatorCfg config.ValidatorConfig, continueOnError bool) error {
	// Create validator
	v := validator.NewValidator(validator.Config{
		Enabled:        validatorCfg.Enabled,
		FailOnWarnings: validatorCfg.FailOnWarnings,
		StrictMode:     validatorCfg.StrictMode,
		CustomRules:    validatorCfg.CustomRules,
		IgnoredRules:   validatorCfg.IgnoredRules,
	})

	// Validate all specs
	results, err := validator.ValidateMultiple(v, specs)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Check for validation failures
	hasErrors := false
	for _, result := range results {
		if !result.Valid {
			hasErrors = true
			// Log detailed validation results with enhanced formatting
			log.Printf("\n%s", validator.FormatValidationResultEnhanced(result))
		} else if len(result.Warnings) > 0 {
			// Use enhanced formatting for warnings too
			log.Printf("\n%s", validator.FormatValidationResultEnhanced(result))
		}
	}

	if hasErrors {
		if !continueOnError {
			return fmt.Errorf("validation failed for one or more specs (see detailed errors above)")
		}
		log.Printf("⚠️  Warning: Some specs failed validation but continuing due to continue_on_error=true")
	}

	return nil
}
