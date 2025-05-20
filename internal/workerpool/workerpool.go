package workerpool

import (
	"context"
	"sync"
)

// Task represents a function that can be executed by a worker
type Task func() interface{}

// Result represents the result of a task execution
type Result struct {
	Value interface{}
	Err   error
}

// WorkerPool manages a pool of workers for concurrent task execution
type WorkerPool struct {
	numWorkers int
	tasks      chan Task
	results    chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// New creates a new worker pool with the specified number of workers
func New(numWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	wp := &WorkerPool{
		numWorkers: numWorkers,
		tasks:      make(chan Task, numWorkers*10), // Buffer tasks channel to avoid blocking
		results:    make(chan Result, numWorkers*10), // Buffer results channel to avoid blocking
		ctx:        ctx,
		cancel:     cancel,
	}
	
	wp.start()
	
	return wp
}

// start launches the worker goroutines
func (wp *WorkerPool) start() {
	// Start the workers
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			
			for {
				select {
				case <-wp.ctx.Done():
					// Context canceled, exit worker
					return
				case task, ok := <-wp.tasks:
					if !ok {
						// Channel closed, exit worker
						return
					}
					
					// Execute the task
					result := task()
					
					// Send the result
					select {
					case <-wp.ctx.Done():
						// Context canceled, don't send result
						return
					case wp.results <- Result{Value: result}:
						// Result sent successfully
					}
				}
			}
		}(i)
	}
}

// Submit adds a task to the worker pool and returns a channel that will receive the result
func (wp *WorkerPool) Submit(task Task) <-chan Result {
	resultCh := make(chan Result, 1)
	
	// Wrap the task to capture its result
	wrappedTask := func() interface{} {
		return task()
	}
	
	// Submit the task
	select {
	case <-wp.ctx.Done():
		// Pool is shutting down, return empty result
		close(resultCh)
	case wp.tasks <- wrappedTask:
		// Wait for the result in a separate goroutine
		go func() {
			result := <-wp.results
			resultCh <- result
			close(resultCh)
		}()
	}
	
	return resultCh
}

// SubmitBatch submits multiple tasks to the worker pool and returns a channel that will receive all results
func (wp *WorkerPool) SubmitBatch(tasks []Task) <-chan Result {
	resultCh := make(chan Result, len(tasks))
	
	// Create a wait group to wait for all tasks to complete
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	
	// Submit each task
	for _, task := range tasks {
		// Capture the task variable
		t := task
		
		// Submit the task to the worker pool
		go func() {
			defer wg.Done()
			
			// Create a wrapped task that returns the result
			wrappedTask := func() interface{} {
				return t()
			}
			
			// Submit the task
			select {
			case <-wp.ctx.Done():
				// Pool is shutting down, skip this task
				return
			case wp.tasks <- wrappedTask:
				// Task submitted, wait for result
				select {
				case <-wp.ctx.Done():
					// Pool is shutting down, skip this task
					return
				case result := <-wp.results:
					// Send the result to the result channel
					resultCh <- result
				}
			}
		}()
	}
	
	// Wait for all tasks to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	
	return resultCh
}

// Shutdown gracefully shuts down the worker pool
// It stops accepting new tasks and waits for all pending tasks to complete
func (wp *WorkerPool) Shutdown() {
	// Signal workers to stop
	wp.cancel()
	
	// Wait for all workers to exit
	wp.wg.Wait()
}

// ShutdownNow immediately shuts down the worker pool
// It stops accepting new tasks and cancels all pending tasks
func (wp *WorkerPool) ShutdownNow() {
	// Signal workers to stop
	wp.cancel()
	
	// Clear the tasks channel
	for len(wp.tasks) > 0 {
		<-wp.tasks
	}
	
	// Wait for all workers to exit
	wp.wg.Wait()
}
