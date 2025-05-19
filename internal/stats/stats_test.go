package stats

import (
	"strings"
	"testing"
	"time"
)

func TestServerStats(t *testing.T) {
	// Create a new ServerStats instance
	stats := NewServerStats()
	
	// Test initial values
	if stats.RequestsProcessed != 0 {
		t.Errorf("Initial RequestsProcessed = %d, want 0", stats.RequestsProcessed)
	}
	
	if stats.MemoryUsed != 0 {
		t.Errorf("Initial MemoryUsed = %d, want 0", stats.MemoryUsed)
	}
	
	if stats.CapacityRatio != 0.0 {
		t.Errorf("Initial CapacityRatio = %f, want 0.0", stats.CapacityRatio)
	}
	
	// Test IncrementRequests
	for i := 0; i < 10; i++ {
		stats.IncrementRequests()
	}
	
	if stats.RequestsProcessed != 10 {
		t.Errorf("After 10 increments, RequestsProcessed = %d, want 10", stats.RequestsProcessed)
	}
	
	// Test concurrent counter
	stats.IncConcurrent()
	stats.IncConcurrent()
	
	if stats.CurrentConcurrent != 2 {
		t.Errorf("CurrentConcurrent = %d, want 2", stats.CurrentConcurrent)
	}
	
	stats.DecConcurrent()
	
	if stats.CurrentConcurrent != 1 {
		t.Errorf("After decrement, CurrentConcurrent = %d, want 1", stats.CurrentConcurrent)
	}
	
	// Test UpdateMemoryUsage
	stats.UpdateMemoryUsage()
	
	// Memory usage should be non-zero after update
	if stats.MemoryUsed == 0 {
		t.Error("After UpdateMemoryUsage, MemoryUsed should be non-zero")
	}
	
	// Test GetStatsReport
	report := stats.GetStatsReport()
	
	// Check that the report contains the expected sections
	if !strings.Contains(report, "requests_processed") {
		t.Error("Stats report does not contain 'requests_processed'")
	}
	
	if !strings.Contains(report, "memory used") {
		t.Error("Stats report does not contain 'memory used'")
	}
	
	if !strings.Contains(report, "capacity of this server") {
		t.Error("Stats report does not contain 'capacity of this server'")
	}
	
	// Test StartMonitoring
	done := make(chan bool)
	
	// Get initial memory value
	initialMemory := stats.MemoryUsed
	
	// Start monitoring with a short interval
	stats.StartMonitoring(50 * time.Millisecond)
	
	// Allocate some memory to make sure the usage changes
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	
	// Wait for a few monitor cycles
	time.Sleep(200 * time.Millisecond)
	
	// Memory usage should have been updated
	if stats.MemoryUsed <= initialMemory {
		t.Error("Memory usage should have increased after allocation")
	}
	
	// Just to make sure data isn't optimized away
	if len(data) == 0 {
		t.Error("Data should not be empty")
	}
	
	close(done)
}
