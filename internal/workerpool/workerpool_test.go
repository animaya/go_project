package workerpool

import (
	"sync"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	// Create a worker pool with 4 workers
	wp := New(4)
	defer wp.Shutdown()
	
	// Test submitting a single task
	t.Run("SubmitSingleTask", func(t *testing.T) {
		// Create a simple task that returns a value
		task := func() interface{} {
			return 42
		}
		
		// Submit the task to the worker pool
		resultCh := wp.Submit(task)
		
		// Get the result
		result := <-resultCh
		
		// Check the result
		if result.Value != 42 {
			t.Errorf("Expected result to be 42, got %v", result.Value)
		}
		if result.Err != nil {
			t.Errorf("Expected error to be nil, got %v", result.Err)
		}
	})
	
	// Test submitting multiple tasks
	t.Run("SubmitMultipleTasks", func(t *testing.T) {
		// Define the number of tasks
		numTasks := 100
		
		// Create tasks
		tasks := make([]Task, numTasks)
		for i := 0; i < numTasks; i++ {
			taskNum := i
			tasks[i] = func() interface{} {
				return taskNum
			}
		}
		
		// Submit tasks in batches
		resultCh := wp.SubmitBatch(tasks)
		
		// Collect results
		results := make(map[int]bool)
		resultsLock := sync.Mutex{}
		
		// Process results as they come in
		for result := range resultCh {
			resultsLock.Lock()
			taskNum, ok := result.Value.(int)
			if !ok {
				t.Errorf("Expected result value to be an int, got %T", result.Value)
			} else {
				results[taskNum] = true
			}
			resultsLock.Unlock()
		}
		
		// Verify all tasks completed
		if len(results) != numTasks {
			t.Errorf("Expected %d results, got %d", numTasks, len(results))
		}
		
		// Verify all task numbers were received
		for i := 0; i < numTasks; i++ {
			if !results[i] {
				t.Errorf("Missing result for task %d", i)
			}
		}
	})
	
	// Test graceful shutdown
	t.Run("GracefulShutdown", func(t *testing.T) {
		// Create a new worker pool for this test
		shutdownWP := New(4)
		
		// Create a channel to signal task completion
		doneCh := make(chan struct{})
		
		// Submit a long-running task
		shutdownWP.Submit(func() interface{} {
			time.Sleep(100 * time.Millisecond)
			doneCh <- struct{}{}
			return nil
		})
		
		// Shutdown the worker pool
		go func() {
			time.Sleep(10 * time.Millisecond)
			shutdownWP.Shutdown()
		}()
		
		// Wait for the task to complete
		select {
		case <-doneCh:
			// Task completed successfully
		case <-time.After(200 * time.Millisecond):
			t.Error("Task did not complete before timeout")
		}
	})
	
	// Test immediate shutdown
	t.Run("ImmediateShutdown", func(t *testing.T) {
		// Create a new worker pool for this test
		shutdownWP := New(4)
		
		// Submit a long-running task
		shutdownWP.Submit(func() interface{} {
			time.Sleep(1 * time.Second)
			return nil
		})
		
		// Immediately shutdown the worker pool
		start := time.Now()
		shutdownWP.ShutdownNow()
		duration := time.Since(start)
		
		// Shutdown should return quickly, not wait for tasks to complete
		if duration > 100*time.Millisecond {
			t.Errorf("ShutdownNow took too long: %v", duration)
		}
	})
}
