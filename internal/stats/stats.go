package stats

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ServerStats holds the server statistics
type ServerStats struct {
	RequestsProcessed uint64
	MemoryUsed        uint64
	CapacityRatio     float64
	StartTime         time.Time
	MaxConcurrent     int64
	CurrentConcurrent int64
	mu                sync.RWMutex
}

// NewServerStats initializes a new ServerStats instance
func NewServerStats() *ServerStats {
	return &ServerStats{
		RequestsProcessed: 0,
		MemoryUsed:        0,
		CapacityRatio:     0.0,
		StartTime:         time.Now(),
		MaxConcurrent:     1000, // Default max concurrent requests
		CurrentConcurrent: 0,
	}
}

// IncrementRequests increments the number of processed requests
func (s *ServerStats) IncrementRequests() {
	atomic.AddUint64(&s.RequestsProcessed, 1)
}

// IncConcurrent increments the current concurrent requests counter
func (s *ServerStats) IncConcurrent() {
	atomic.AddInt64(&s.CurrentConcurrent, 1)
}

// DecConcurrent decrements the current concurrent requests counter
func (s *ServerStats) DecConcurrent() {
	atomic.AddInt64(&s.CurrentConcurrent, -1)
}

// UpdateMemoryUsage updates the memory usage statistics
func (s *ServerStats) UpdateMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	atomic.StoreUint64(&s.MemoryUsed, m.Alloc)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Calculate capacity ratio (current / max)
	current := float64(atomic.LoadInt64(&s.CurrentConcurrent))
	max := float64(s.MaxConcurrent)
	s.CapacityRatio = current / max
}

// GetStatsReport returns a formatted string with the server statistics
func (s *ServerStats) GetStatsReport() string {
	s.UpdateMemoryUsage()
	
	return fmt.Sprintf(`## Web server statistics
### requests_processed - %d
### memory used - %.2f MB
### capacity of this server for processing requests - %.1f/10`,
		atomic.LoadUint64(&s.RequestsProcessed),
		float64(atomic.LoadUint64(&s.MemoryUsed))/1024/1024,
		s.CapacityRatio*10)
}

// StartMonitoring starts a goroutine that periodically updates the stats
func (s *ServerStats) StartMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			<-ticker.C
			s.UpdateMemoryUsage()
		}
	}()
}
