package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Task represents a unit of work to be processed by the worker pool
type Task struct {
	ID      string
	Execute func(ctx context.Context) error
}

// Result represents the result of processing a task
type Result struct {
	TaskID string
	Error  error
}

// Pool manages a pool of workers for concurrent task execution
type Pool struct {
	workerCount int
	tasks       chan Task
	results     chan Result
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	started     bool
}

// Config contains configuration for the worker pool
type Config struct {
	// Number of workers in the pool (defaults to 4)
	WorkerCount int
	// Buffer size for task queue (defaults to 100)
	TaskQueueSize int
}

// NewPool creates a new worker pool with the given configuration
func NewPool(cfg Config) *Pool {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	if cfg.TaskQueueSize <= 0 {
		cfg.TaskQueueSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workerCount: cfg.WorkerCount,
		tasks:       make(chan Task, cfg.TaskQueueSize),
		results:     make(chan Result, cfg.TaskQueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start initializes and starts all workers in the pool
func (p *Pool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("pool already started")
	}

	log.Printf("Starting worker pool with %d workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i + 1)
	}

	p.started = true
	return nil
}

// worker is the worker goroutine that processes tasks from the queue
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker %d stopping due to context cancellation", id)
			return

		case task, ok := <-p.tasks:
			if !ok {
				log.Printf("Worker %d stopping: task channel closed", id)
				return
			}

			log.Printf("Worker %d processing task: %s", id, task.ID)

			// Execute the task
			err := task.Execute(p.ctx)

			// Send result
			select {
			case p.results <- Result{TaskID: task.ID, Error: err}:
				if err != nil {
					log.Printf("Worker %d completed task %s with error: %v", id, task.ID, err)
				} else {
					log.Printf("Worker %d completed task %s successfully", id, task.ID)
				}
			case <-p.ctx.Done():
				log.Printf("Worker %d unable to send result: context cancelled", id)
				return
			}
		}
	}
}

// Submit adds a task to the pool's queue
func (p *Pool) Submit(task Task) error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return fmt.Errorf("pool not started")
	}
	p.mu.Unlock()

	select {
	case p.tasks <- task:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("pool context cancelled")
	}
}

// Wait closes the task channel and waits for all workers to complete
// Returns all results collected from workers
func (p *Pool) Wait() []Result {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return []Result{}
	}
	p.mu.Unlock()

	// Close task channel to signal no more tasks
	close(p.tasks)

	// Wait for all workers to finish
	p.wg.Wait()

	// Close results channel
	close(p.results)

	// Collect all results
	var results []Result
	for result := range p.results {
		results = append(results, result)
	}

	return results
}

// Shutdown cancels the pool context and waits for all workers to stop
func (p *Pool) Shutdown() {
	log.Printf("Shutting down worker pool")
	p.cancel()
	p.wg.Wait()
}

// ProcessBatch submits multiple tasks and waits for all to complete
// Returns results for all tasks in the order they complete
func (p *Pool) ProcessBatch(ctx context.Context, tasks []Task) ([]Result, error) {
	// Start the pool if not already started
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		if err := p.Start(); err != nil {
			return nil, fmt.Errorf("failed to start pool: %w", err)
		}
	} else {
		p.mu.Unlock()
	}

	// Submit all tasks
	for _, task := range tasks {
		if err := p.Submit(task); err != nil {
			return nil, fmt.Errorf("failed to submit task %s: %w", task.ID, err)
		}
	}

	// Wait for results with context cancellation support
	resultsChan := make(chan []Result, 1)
	go func() {
		resultsChan <- p.Wait()
	}()

	select {
	case results := <-resultsChan:
		return results, nil
	case <-ctx.Done():
		p.Shutdown()
		return nil, fmt.Errorf("batch processing cancelled: %w", ctx.Err())
	}
}
