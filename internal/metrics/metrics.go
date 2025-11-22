package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Metrics holds aggregated generation metrics
type Metrics struct {
	mu                sync.RWMutex
	TotalSpecs        int             `json:"total_specs"`
	SuccessfulSpecs   int             `json:"successful_specs"`
	FailedSpecs       int             `json:"failed_specs"`
	CachedSpecs       int             `json:"cached_specs"`
	TotalDurationMs   int64           `json:"total_duration_ms"`
	AverageDurationMs int64           `json:"average_duration_ms"`
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time"`
	SpecMetrics       []SpecMetric    `json:"spec_metrics"`
}

// SpecMetric holds metrics for a single spec generation
type SpecMetric struct {
	SpecPath      string    `json:"spec_path"`
	ServiceName   string    `json:"service_name"`
	Success       bool      `json:"success"`
	Cached        bool      `json:"cached"`
	DurationMs    int64     `json:"duration_ms"`
	Error         string    `json:"error,omitempty"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// Collector collects metrics during generation
type Collector struct {
	metrics *Metrics
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics: &Metrics{
			SpecMetrics: make([]SpecMetric, 0),
			StartTime:   time.Now(),
		},
	}
}

// RecordSpec records metrics for a single spec generation
func (c *Collector) RecordSpec(metric SpecMetric) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.TotalSpecs++
	if metric.Success {
		c.metrics.SuccessfulSpecs++
	} else {
		c.metrics.FailedSpecs++
	}
	if metric.Cached {
		c.metrics.CachedSpecs++
	}

	c.metrics.TotalDurationMs += metric.DurationMs
	c.metrics.SpecMetrics = append(c.metrics.SpecMetrics, metric)
}

// Finalize calculates final metrics before export
func (c *Collector) Finalize() {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.EndTime = time.Now()
	if c.metrics.TotalSpecs > 0 {
		c.metrics.AverageDurationMs = c.metrics.TotalDurationMs / int64(c.metrics.TotalSpecs)
	}
}

// Export exports metrics to a JSON file
func (c *Collector) Export(path string) error {
	c.Finalize()

	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	data, err := json.MarshalIndent(c.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// Summary returns a human-readable summary
func (c *Collector) Summary() string {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	totalSecs := c.metrics.TotalDurationMs / 1000
	avgSecs := c.metrics.AverageDurationMs / 1000

	return fmt.Sprintf(
		"Generation Summary: %d total, %d successful, %d failed, %d cached (%.1fs total, %.1fs avg)",
		c.metrics.TotalSpecs,
		c.metrics.SuccessfulSpecs,
		c.metrics.FailedSpecs,
		c.metrics.CachedSpecs,
		float64(totalSecs),
		float64(avgSecs),
	)
}

// GetMetrics returns a copy of the current metrics (safe for concurrent access)
func (c *Collector) GetMetrics() Metrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	metricsCopy := *c.metrics
	metricsCopy.SpecMetrics = make([]SpecMetric, len(c.metrics.SpecMetrics))
	copy(metricsCopy.SpecMetrics, c.metrics.SpecMetrics)

	return metricsCopy
}

// SuccessRate returns the success rate as a percentage
func (c *Collector) SuccessRate() float64 {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	if c.metrics.TotalSpecs == 0 {
		return 0.0
	}
	return float64(c.metrics.SuccessfulSpecs) / float64(c.metrics.TotalSpecs) * 100.0
}

// CacheHitRate returns the cache hit rate as a percentage
func (c *Collector) CacheHitRate() float64 {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	if c.metrics.TotalSpecs == 0 {
		return 0.0
	}
	return float64(c.metrics.CachedSpecs) / float64(c.metrics.TotalSpecs) * 100.0
}
