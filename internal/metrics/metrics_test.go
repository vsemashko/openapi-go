package metrics

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()
	if collector == nil {
		t.Fatal("Expected collector to be created")
	}
	if collector.metrics == nil {
		t.Fatal("Expected metrics to be initialized")
	}
	if collector.metrics.SpecMetrics == nil {
		t.Fatal("Expected spec metrics slice to be initialized")
	}
}

func TestRecordSpec(t *testing.T) {
	collector := NewCollector()

	metric := SpecMetric{
		SpecPath:    "/path/to/spec.json",
		ServiceName: "test-service",
		Success:     true,
		Cached:      false,
		DurationMs:  1000,
		GeneratedAt: time.Now(),
	}

	collector.RecordSpec(metric)

	metrics := collector.GetMetrics()
	if metrics.TotalSpecs != 1 {
		t.Errorf("Expected TotalSpecs=1, got %d", metrics.TotalSpecs)
	}
	if metrics.SuccessfulSpecs != 1 {
		t.Errorf("Expected SuccessfulSpecs=1, got %d", metrics.SuccessfulSpecs)
	}
	if metrics.FailedSpecs != 0 {
		t.Errorf("Expected FailedSpecs=0, got %d", metrics.FailedSpecs)
	}
	if metrics.CachedSpecs != 0 {
		t.Errorf("Expected CachedSpecs=0, got %d", metrics.CachedSpecs)
	}
	if metrics.TotalDurationMs != 1000 {
		t.Errorf("Expected TotalDurationMs=1000, got %d", metrics.TotalDurationMs)
	}
}

func TestRecordMultipleSpecs(t *testing.T) {
	collector := NewCollector()

	// Record 3 successful, 1 failed, 2 cached
	specs := []SpecMetric{
		{SpecPath: "/spec1.json", ServiceName: "svc1", Success: true, Cached: false, DurationMs: 1000},
		{SpecPath: "/spec2.json", ServiceName: "svc2", Success: true, Cached: true, DurationMs: 100},
		{SpecPath: "/spec3.json", ServiceName: "svc3", Success: false, Cached: false, DurationMs: 500, Error: "test error"},
		{SpecPath: "/spec4.json", ServiceName: "svc4", Success: true, Cached: true, DurationMs: 50},
	}

	for _, spec := range specs {
		collector.RecordSpec(spec)
	}

	metrics := collector.GetMetrics()
	if metrics.TotalSpecs != 4 {
		t.Errorf("Expected TotalSpecs=4, got %d", metrics.TotalSpecs)
	}
	if metrics.SuccessfulSpecs != 3 {
		t.Errorf("Expected SuccessfulSpecs=3, got %d", metrics.SuccessfulSpecs)
	}
	if metrics.FailedSpecs != 1 {
		t.Errorf("Expected FailedSpecs=1, got %d", metrics.FailedSpecs)
	}
	if metrics.CachedSpecs != 2 {
		t.Errorf("Expected CachedSpecs=2, got %d", metrics.CachedSpecs)
	}
	if metrics.TotalDurationMs != 1650 {
		t.Errorf("Expected TotalDurationMs=1650, got %d", metrics.TotalDurationMs)
	}
}

func TestFinalize(t *testing.T) {
	collector := NewCollector()

	collector.RecordSpec(SpecMetric{Success: true, DurationMs: 1000})
	collector.RecordSpec(SpecMetric{Success: true, DurationMs: 2000})

	collector.Finalize()

	metrics := collector.GetMetrics()
	if metrics.AverageDurationMs != 1500 {
		t.Errorf("Expected AverageDurationMs=1500, got %d", metrics.AverageDurationMs)
	}
	if metrics.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}
}

func TestExport(t *testing.T) {
	collector := NewCollector()

	collector.RecordSpec(SpecMetric{
		SpecPath:    "/spec.json",
		ServiceName: "test-service",
		Success:     true,
		Cached:      false,
		DurationMs:  1000,
		GeneratedAt: time.Now(),
	})

	tmpFile := t.TempDir() + "/metrics.json"
	err := collector.Export(tmpFile)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected metrics file to exist")
	}

	// Verify JSON is valid
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read metrics file: %v", err)
	}

	var metrics Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("Failed to parse metrics JSON: %v", err)
	}

	if metrics.TotalSpecs != 1 {
		t.Errorf("Expected TotalSpecs=1 in exported file, got %d", metrics.TotalSpecs)
	}
}

func TestSummary(t *testing.T) {
	collector := NewCollector()

	collector.RecordSpec(SpecMetric{Success: true, Cached: false, DurationMs: 1000})
	collector.RecordSpec(SpecMetric{Success: true, Cached: true, DurationMs: 100})
	collector.RecordSpec(SpecMetric{Success: false, Cached: false, DurationMs: 500})

	summary := collector.Summary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Summary should contain key information
	expectedParts := []string{"3 total", "2 successful", "1 failed", "1 cached"}
	for _, part := range expectedParts {
		if !contains(summary, part) {
			t.Errorf("Expected summary to contain '%s', got: %s", part, summary)
		}
	}
}

func TestSuccessRate(t *testing.T) {
	tests := []struct {
		name         string
		specs        []SpecMetric
		expectedRate float64
	}{
		{
			name:         "no specs",
			specs:        []SpecMetric{},
			expectedRate: 0.0,
		},
		{
			name: "all successful",
			specs: []SpecMetric{
				{Success: true},
				{Success: true},
			},
			expectedRate: 100.0,
		},
		{
			name: "50% successful",
			specs: []SpecMetric{
				{Success: true},
				{Success: false},
			},
			expectedRate: 50.0,
		},
		{
			name: "all failed",
			specs: []SpecMetric{
				{Success: false},
				{Success: false},
			},
			expectedRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCollector()
			for _, spec := range tt.specs {
				collector.RecordSpec(spec)
			}

			rate := collector.SuccessRate()
			if rate != tt.expectedRate {
				t.Errorf("Expected success rate %.1f%%, got %.1f%%", tt.expectedRate, rate)
			}
		})
	}
}

func TestCacheHitRate(t *testing.T) {
	tests := []struct {
		name         string
		specs        []SpecMetric
		expectedRate float64
	}{
		{
			name:         "no specs",
			specs:        []SpecMetric{},
			expectedRate: 0.0,
		},
		{
			name: "all cached",
			specs: []SpecMetric{
				{Cached: true},
				{Cached: true},
			},
			expectedRate: 100.0,
		},
		{
			name: "50% cached",
			specs: []SpecMetric{
				{Cached: true},
				{Cached: false},
			},
			expectedRate: 50.0,
		},
		{
			name: "none cached",
			specs: []SpecMetric{
				{Cached: false},
				{Cached: false},
			},
			expectedRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCollector()
			for _, spec := range tt.specs {
				collector.RecordSpec(spec)
			}

			rate := collector.CacheHitRate()
			if rate != tt.expectedRate {
				t.Errorf("Expected cache hit rate %.1f%%, got %.1f%%", tt.expectedRate, rate)
			}
		})
	}
}

func TestConcurrentRecording(t *testing.T) {
	collector := NewCollector()

	// Simulate concurrent recording
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				collector.RecordSpec(SpecMetric{
					Success:    j%2 == 0,
					DurationMs: int64(j),
				})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := collector.GetMetrics()
	if metrics.TotalSpecs != 1000 {
		t.Errorf("Expected TotalSpecs=1000 from concurrent recording, got %d", metrics.TotalSpecs)
	}
}

func TestGetMetricsCopy(t *testing.T) {
	collector := NewCollector()
	collector.RecordSpec(SpecMetric{SpecPath: "/spec1.json", Success: true})

	// Get first copy
	metrics1 := collector.GetMetrics()

	// Record another spec
	collector.RecordSpec(SpecMetric{SpecPath: "/spec2.json", Success: true})

	// Get second copy
	metrics2 := collector.GetMetrics()

	// First copy should be unchanged
	if metrics1.TotalSpecs != 1 {
		t.Errorf("Expected first copy TotalSpecs=1, got %d", metrics1.TotalSpecs)
	}

	// Second copy should have both
	if metrics2.TotalSpecs != 2 {
		t.Errorf("Expected second copy TotalSpecs=2, got %d", metrics2.TotalSpecs)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
