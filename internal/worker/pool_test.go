package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name                string
		config              Config
		expectedWorkerCount int
		expectedQueueSize   int
	}{
		{
			name:                "default config",
			config:              Config{},
			expectedWorkerCount: 4,
			expectedQueueSize:   100,
		},
		{
			name: "custom config",
			config: Config{
				WorkerCount:   8,
				TaskQueueSize: 50,
			},
			expectedWorkerCount: 8,
			expectedQueueSize:   50,
		},
		{
			name: "zero values use defaults",
			config: Config{
				WorkerCount:   0,
				TaskQueueSize: 0,
			},
			expectedWorkerCount: 4,
			expectedQueueSize:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(tt.config)

			if pool == nil {
				t.Fatal("NewPool() returned nil")
			}

			if pool.workerCount != tt.expectedWorkerCount {
				t.Errorf("workerCount = %d, want %d", pool.workerCount, tt.expectedWorkerCount)
			}

			if cap(pool.tasks) != tt.expectedQueueSize {
				t.Errorf("task queue size = %d, want %d", cap(pool.tasks), tt.expectedQueueSize)
			}

			if cap(pool.results) != tt.expectedQueueSize {
				t.Errorf("results queue size = %d, want %d", cap(pool.results), tt.expectedQueueSize)
			}
		})
	}
}

func TestPoolStart(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 2})

	// First start should succeed
	if err := pool.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Second start should fail
	if err := pool.Start(); err == nil {
		t.Error("Start() should fail when pool already started")
	}

	pool.Shutdown()
}

func TestPoolSubmit(t *testing.T) {
	tests := []struct {
		name      string
		startPool bool
		wantErr   bool
	}{
		{
			name:      "submit to started pool",
			startPool: true,
			wantErr:   false,
		},
		{
			name:      "submit to non-started pool",
			startPool: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(Config{WorkerCount: 2})

			if tt.startPool {
				pool.Start()
				defer pool.Shutdown()
			}

			task := Task{
				ID: "test-task",
				Execute: func(ctx context.Context) error {
					return nil
				},
			}

			err := pool.Submit(task)

			if (err != nil) != tt.wantErr {
				t.Errorf("Submit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPoolProcessBatch(t *testing.T) {
	tests := []struct {
		name      string
		taskCount int
		taskFunc  func(int) func(context.Context) error
		wantErr   bool
	}{
		{
			name:      "successful batch",
			taskCount: 10,
			taskFunc: func(i int) func(context.Context) error {
				return func(ctx context.Context) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "batch with errors",
			taskCount: 5,
			taskFunc: func(i int) func(context.Context) error {
				return func(ctx context.Context) error {
					if i%2 == 0 {
						return fmt.Errorf("task %d error", i)
					}
					return nil
				}
			},
			wantErr: false, // ProcessBatch itself shouldn't error, but results will contain errors
		},
		{
			name:      "empty batch",
			taskCount: 0,
			taskFunc:  nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(Config{WorkerCount: 4})

			var tasks []Task
			for i := 0; i < tt.taskCount; i++ {
				task := Task{
					ID: fmt.Sprintf("task-%d", i),
				}
				if tt.taskFunc != nil {
					task.Execute = tt.taskFunc(i)
				} else {
					task.Execute = func(ctx context.Context) error { return nil }
				}
				tasks = append(tasks, task)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			results, err := pool.ProcessBatch(ctx, tasks)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && len(results) != tt.taskCount {
				t.Errorf("ProcessBatch() returned %d results, want %d", len(results), tt.taskCount)
			}

			// Count errors in results
			var errorCount int
			for _, result := range results {
				if result.Error != nil {
					errorCount++
				}
			}

			t.Logf("Processed %d tasks, %d errors", len(results), errorCount)
		})
	}
}

func TestPoolCancellation(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 2})

	var started atomic.Int32
	var completed atomic.Int32

	tasks := []Task{}
	for i := 0; i < 10; i++ {
		task := Task{
			ID: fmt.Sprintf("task-%d", i),
			Execute: func(ctx context.Context) error {
				started.Add(1)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
					completed.Add(1)
					return nil
				}
			},
		}
		tasks = append(tasks, task)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := pool.ProcessBatch(ctx, tasks)

	if err == nil {
		t.Error("ProcessBatch() should return error when context is cancelled")
	}

	t.Logf("Started: %d, Completed: %d tasks", started.Load(), completed.Load())

	// Verify not all tasks completed (due to cancellation)
	if completed.Load() == int32(len(tasks)) {
		t.Error("All tasks completed despite cancellation")
	}
}

func TestPoolConcurrency(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 4})

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32
	var mu sync.Mutex

	updateMax := func(current int32) {
		mu.Lock()
		defer mu.Unlock()
		if current > maxConcurrent.Load() {
			maxConcurrent.Store(current)
		}
	}

	tasks := []Task{}
	for i := 0; i < 20; i++ {
		task := Task{
			ID: fmt.Sprintf("task-%d", i),
			Execute: func(ctx context.Context) error {
				current := concurrent.Add(1)
				updateMax(current)
				time.Sleep(50 * time.Millisecond)
				concurrent.Add(-1)
				return nil
			},
		}
		tasks = append(tasks, task)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := pool.ProcessBatch(ctx, tasks)

	if err != nil {
		t.Fatalf("ProcessBatch() failed: %v", err)
	}

	if len(results) != len(tasks) {
		t.Errorf("ProcessBatch() returned %d results, want %d", len(results), len(tasks))
	}

	// Verify we actually ran tasks concurrently
	maxC := maxConcurrent.Load()
	if maxC < 2 {
		t.Errorf("maxConcurrent = %d, expected at least 2 (indicating parallel execution)", maxC)
	}

	// Should not exceed worker count
	if maxC > 4 {
		t.Errorf("maxConcurrent = %d, expected at most 4 (worker count)", maxC)
	}

	t.Logf("Max concurrent tasks: %d (worker count: 4)", maxC)
}

func TestPoolErrorHandling(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 2})

	tasks := []Task{
		{
			ID: "success-1",
			Execute: func(ctx context.Context) error {
				return nil
			},
		},
		{
			ID: "error-1",
			Execute: func(ctx context.Context) error {
				return fmt.Errorf("intentional error")
			},
		},
		{
			ID: "success-2",
			Execute: func(ctx context.Context) error {
				return nil
			},
		},
		{
			ID: "error-2",
			Execute: func(ctx context.Context) error {
				return fmt.Errorf("another error")
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := pool.ProcessBatch(ctx, tasks)

	if err != nil {
		t.Fatalf("ProcessBatch() failed: %v", err)
	}

	if len(results) != len(tasks) {
		t.Errorf("ProcessBatch() returned %d results, want %d", len(results), len(tasks))
	}

	// Count successes and errors
	var successCount, errorCount int
	resultMap := make(map[string]error)

	for _, result := range results {
		resultMap[result.TaskID] = result.Error
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successful tasks, got %d", successCount)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 failed tasks, got %d", errorCount)
	}

	// Verify specific task results
	if resultMap["success-1"] != nil {
		t.Errorf("Task success-1 should succeed, got error: %v", resultMap["success-1"])
	}

	if resultMap["error-1"] == nil {
		t.Error("Task error-1 should fail")
	}
}

func TestPoolShutdown(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 2})

	if err := pool.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Submit a long-running task
	task := Task{
		ID: "long-task",
		Execute: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
	}

	if err := pool.Submit(task); err != nil {
		t.Fatalf("Submit() failed: %v", err)
	}

	// Shutdown should cancel the context and wait for workers
	done := make(chan struct{})
	go func() {
		pool.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Shutdown completed successfully")
	case <-time.After(2 * time.Second):
		t.Error("Shutdown() took too long, should cancel tasks quickly")
	}
}

func TestPoolWait(t *testing.T) {
	pool := NewPool(Config{WorkerCount: 2})

	if err := pool.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Submit tasks
	taskCount := 5
	for i := 0; i < taskCount; i++ {
		task := Task{
			ID: fmt.Sprintf("task-%d", i),
			Execute: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}
		if err := pool.Submit(task); err != nil {
			t.Fatalf("Submit() failed: %v", err)
		}
	}

	// Wait should return all results
	results := pool.Wait()

	if len(results) != taskCount {
		t.Errorf("Wait() returned %d results, want %d", len(results), taskCount)
	}
}

func TestPoolRaceConditions(t *testing.T) {
	// This test is designed to be run with -race flag
	// Each goroutine uses its own pool instance since pools are single-use

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// Concurrent batch processing with separate pools
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create a new pool for this goroutine
			pool := NewPool(Config{WorkerCount: 4})

			tasks := []Task{}
			for j := 0; j < 10; j++ {
				task := Task{
					ID: fmt.Sprintf("task-%d-%d", id, j),
					Execute: func(ctx context.Context) error {
						time.Sleep(time.Millisecond)
						return nil
					},
				}
				tasks = append(tasks, task)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := pool.ProcessBatch(ctx, tasks)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("Concurrent ProcessBatch() error: %v", err)
	}
}
