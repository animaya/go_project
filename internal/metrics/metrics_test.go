package metrics

import (
	"errors"
	"testing"
	"time"
)

func TestConcurrentTimeSlice(t *testing.T) {
	// Create a new time slice
	timeSlice := NewConcurrentTimeSlice()
	
	// Test initial state
	if timeSlice.Len() != 0 {
		t.Errorf("Expected initial length to be 0, got %d", timeSlice.Len())
	}
	
	if timeSlice.Average() != 0 {
		t.Errorf("Expected initial average to be 0, got %v", timeSlice.Average())
	}
	
	if timeSlice.GetPercentile(50) != 0 {
		t.Errorf("Expected initial P50 to be 0, got %v", timeSlice.GetPercentile(50))
	}
	
	// Add some times
	timeSlice.Add(100 * time.Millisecond)
	timeSlice.Add(200 * time.Millisecond)
	timeSlice.Add(300 * time.Millisecond)
	
	// Test after adding times
	if timeSlice.Len() != 3 {
		t.Errorf("Expected length to be 3, got %d", timeSlice.Len())
	}
	
	if timeSlice.Average() != 200*time.Millisecond {
		t.Errorf("Expected average to be 200ms, got %v", timeSlice.Average())
	}
	
	if timeSlice.GetPercentile(50) != 200*time.Millisecond {
		t.Errorf("Expected P50 to be 200ms, got %v", timeSlice.GetPercentile(50))
	}
	
	// Test other percentiles
	if timeSlice.GetPercentile(0) != 100*time.Millisecond {
		t.Errorf("Expected P0 to be 100ms, got %v", timeSlice.GetPercentile(0))
	}
	
	if timeSlice.GetPercentile(100) != 300*time.Millisecond {
		t.Errorf("Expected P100 to be 300ms, got %v", timeSlice.GetPercentile(100))
	}
	
	// Add many more times to test the limit
	for i := 0; i < 10001; i++ {
		timeSlice.Add(time.Duration(i) * time.Millisecond)
	}
	
	// The slice should be limited to 10,000 elements
	if timeSlice.Len() != 10000 {
		t.Errorf("Expected length to be 10000, got %d", timeSlice.Len())
	}
	
	// The slice should contain the most recent 10,000 elements (4 through 10003)
	expectedMin := 4 * time.Millisecond
	expectedMax := 10003 * time.Millisecond
	
	if timeSlice.GetPercentile(0) < expectedMin {
		t.Errorf("Expected P0 to be at least %v, got %v", expectedMin, timeSlice.GetPercentile(0))
	}
	
	if timeSlice.GetPercentile(100) > expectedMax {
		t.Errorf("Expected P100 to be at most %v, got %v", expectedMax, timeSlice.GetPercentile(100))
	}
}

func TestMetricsCollector(t *testing.T) {
	// Create a new metrics collector
	collector := NewMetricsCollector(100)
	defer collector.Shutdown()
	
	// Test initial state
	if collector.GetRequestTotal() != 0 {
		t.Errorf("Expected initial request total to be 0, got %d", collector.GetRequestTotal())
	}
	
	if collector.GetRequestSucceeded() != 0 {
		t.Errorf("Expected initial request succeeded to be 0, got %d", collector.GetRequestSucceeded())
	}
	
	if collector.GetRequestFailed() != 0 {
		t.Errorf("Expected initial request failed to be 0, got %d", collector.GetRequestFailed())
	}
	
	if collector.GetCurrentConcurrent() != 0 {
		t.Errorf("Expected initial current concurrent to be 0, got %d", collector.GetCurrentConcurrent())
	}
	
	// Record a successful request
	done := collector.RecordRequest()
	time.Sleep(10 * time.Millisecond) // Simulate request processing
	done(nil)
	
	// Test after recording a successful request
	if collector.GetRequestTotal() != 1 {
		t.Errorf("Expected request total to be 1, got %d", collector.GetRequestTotal())
	}
	
	if collector.GetRequestSucceeded() != 1 {
		t.Errorf("Expected request succeeded to be 1, got %d", collector.GetRequestSucceeded())
	}
	
	if collector.GetRequestFailed() != 0 {
		t.Errorf("Expected request failed to be 0, got %d", collector.GetRequestFailed())
	}
	
	if collector.GetCurrentConcurrent() != 0 {
		t.Errorf("Expected current concurrent to be 0, got %d", collector.GetCurrentConcurrent())
	}
	
	// Record a failed request
	done = collector.RecordRequest()
	time.Sleep(10 * time.Millisecond) // Simulate request processing
	done(errors.New("test error"))
	
	// Test after recording a failed request
	if collector.GetRequestTotal() != 2 {
		t.Errorf("Expected request total to be 2, got %d", collector.GetRequestTotal())
	}
	
	if collector.GetRequestSucceeded() != 1 {
		t.Errorf("Expected request succeeded to be 1, got %d", collector.GetRequestSucceeded())
	}
	
	if collector.GetRequestFailed() != 1 {
		t.Errorf("Expected request failed to be 1, got %d", collector.GetRequestFailed())
	}
	
	// Test concurrent request tracking
	done1 := collector.RecordRequest()
	if collector.GetCurrentConcurrent() != 1 {
		t.Errorf("Expected current concurrent to be 1, got %d", collector.GetCurrentConcurrent())
	}
	
	done2 := collector.RecordRequest()
	if collector.GetCurrentConcurrent() != 2 {
		t.Errorf("Expected current concurrent to be 2, got %d", collector.GetCurrentConcurrent())
	}
	
	done1(nil)
	if collector.GetCurrentConcurrent() != 1 {
		t.Errorf("Expected current concurrent to be 1, got %d", collector.GetCurrentConcurrent())
	}
	
	done2(nil)
	if collector.GetCurrentConcurrent() != 0 {
		t.Errorf("Expected current concurrent to be 0, got %d", collector.GetCurrentConcurrent())
	}
	
	// Test response time tracking
	for i := 0; i < 10; i++ {
		done := collector.RecordRequest()
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
		done(nil)
	}
	
	// Check that response time metrics are reasonable
	metrics := collector.GetCurrentMetrics()
	
	if avgTime, ok := metrics["avg_response_time"].(string); !ok || avgTime == "0s" {
		t.Errorf("Expected average response time to be set, got %v", metrics["avg_response_time"])
	}
	
	if p50Time, ok := metrics["p50_response_time"].(string); !ok || p50Time == "0s" {
		t.Errorf("Expected P50 response time to be set, got %v", metrics["p50_response_time"])
	}
	
	if p90Time, ok := metrics["p90_response_time"].(string); !ok || p90Time == "0s" {
		t.Errorf("Expected P90 response time to be set, got %v", metrics["p90_response_time"])
	}
	
	if p99Time, ok := metrics["p99_response_time"].(string); !ok || p99Time == "0s" {
		t.Errorf("Expected P99 response time to be set, got %v", metrics["p99_response_time"])
	}
	
	// Test GetStatsReport
	report := collector.GetStatsReport()
	if report == "" {
		t.Error("Expected stats report to be non-empty")
	}
	
	// Test individual getters
	if collector.GetUptime() <= 0 {
		t.Error("Expected uptime to be positive")
	}
	
	if collector.GetMemoryUsage() <= 0 {
		t.Error("Expected memory usage to be positive")
	}
	
	if collector.GetCPUUsage() < 0 || collector.GetCPUUsage() > 1 {
		t.Errorf("Expected CPU usage to be between 0 and 1, got %f", collector.GetCPUUsage())
	}
	
	if collector.GetAverageResponseTime() <= 0 {
		t.Error("Expected average response time to be positive")
	}
	
	if collector.GetResponseTimePercentile(50) <= 0 {
		t.Error("Expected P50 response time to be positive")
	}
}
