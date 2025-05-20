package metrics

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects server performance metrics
type MetricsCollector struct {
	startTime         time.Time
	requestsTotal     uint64
	requestsSucceeded uint64
	requestsFailed    uint64
	responseTimes     *ConcurrentTimeSlice
	maxConcurrent     int64
	currentConcurrent int64
	memoryUsage       uint64
	cpuUsage          float64
	mutex             sync.RWMutex
	stopCh            chan struct{}
}

// ConcurrentTimeSlice is a thread-safe slice of response times
type ConcurrentTimeSlice struct {
	times []time.Duration
	mutex sync.RWMutex
}

// NewConcurrentTimeSlice creates a new concurrent time slice
func NewConcurrentTimeSlice() *ConcurrentTimeSlice {
	return &ConcurrentTimeSlice{
		times: make([]time.Duration, 0, 1000), // Pre-allocate for performance
	}
}

// Add adds a new response time to the slice
func (s *ConcurrentTimeSlice) Add(t time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.times = append(s.times, t)
	
	// Limit the size of the slice to prevent memory leaks
	// Keep the most recent 10,000 samples
	if len(s.times) > 10000 {
		s.times = s.times[len(s.times)-10000:]
	}
}

// GetPercentile returns the nth percentile of response times
func (s *ConcurrentTimeSlice) GetPercentile(percentile float64) time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if len(s.times) == 0 {
		return 0
	}
	
	// Make a copy of the times slice to avoid modifying the original
	timesCopy := make([]time.Duration, len(s.times))
	copy(timesCopy, s.times)
	
	// Sort the copy
	sort.Slice(timesCopy, func(i, j int) bool {
		return timesCopy[i] < timesCopy[j]
	})
	
	// Calculate the index for the percentile
	index := int(float64(len(timesCopy)-1) * percentile / 100.0)
	
	return timesCopy[index]
}

// Len returns the number of response times in the slice
func (s *ConcurrentTimeSlice) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.times)
}

// Average returns the average response time
func (s *ConcurrentTimeSlice) Average() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if len(s.times) == 0 {
		return 0
	}
	
	var sum time.Duration
	for _, t := range s.times {
		sum += t
	}
	
	return sum / time.Duration(len(s.times))
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(maxConcurrent int64) *MetricsCollector {
	collector := &MetricsCollector{
		startTime:         time.Now(),
		responseTimes:     NewConcurrentTimeSlice(),
		maxConcurrent:     maxConcurrent,
		currentConcurrent: 0,
		stopCh:            make(chan struct{}),
	}
	
	// Start a goroutine to periodically update system metrics
	go collector.updateSystemMetrics()
	
	return collector
}

// updateSystemMetrics periodically updates system metrics
func (m *MetricsCollector) updateSystemMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.updateMemoryUsage()
			m.updateCPUUsage()
		case <-m.stopCh:
			return
		}
	}
}

// updateMemoryUsage updates the memory usage metric
func (m *MetricsCollector) updateMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	atomic.StoreUint64(&m.memoryUsage, memStats.Alloc)
}

// updateCPUUsage updates the CPU usage metric
// This is a simplified version; in a real system, you would use OS-specific
// APIs to get the actual CPU usage
func (m *MetricsCollector) updateCPUUsage() {
	// Simulate CPU usage based on the number of goroutines
	numGoroutines := runtime.NumGoroutine()
	
	// Calculate a simulated CPU usage based on the number of goroutines
	// and the current concurrent requests
	concurrentRatio := float64(atomic.LoadInt64(&m.currentConcurrent)) / float64(m.maxConcurrent)
	baseUsage := float64(numGoroutines) / 1000.0 // Arbitrary scale
	
	// Combine the metrics with some randomness
	usage := baseUsage*0.3 + concurrentRatio*0.7
	
	// Ensure the usage is between 0 and 1
	if usage > 1.0 {
		usage = 1.0
	}
	
	m.mutex.Lock()
	m.cpuUsage = usage
	m.mutex.Unlock()
}

// RecordRequest records the start of a request and returns a function to call when the request is complete
func (m *MetricsCollector) RecordRequest() func(err error) {
	// Increment the request counter
	atomic.AddUint64(&m.requestsTotal, 1)
	
	// Increment the concurrent requests counter
	atomic.AddInt64(&m.currentConcurrent, 1)
	
	// Record the start time
	startTime := time.Now()
	
	// Return a function to call when the request is complete
	return func(err error) {
		// Record the response time
		responseTime := time.Since(startTime)
		m.responseTimes.Add(responseTime)
		
		// Decrement the concurrent requests counter
		atomic.AddInt64(&m.currentConcurrent, -1)
		
		// Increment the success or failure counter
		if err == nil {
			atomic.AddUint64(&m.requestsSucceeded, 1)
		} else {
			atomic.AddUint64(&m.requestsFailed, 1)
		}
	}
}

// GetCurrentMetrics returns the current metrics
func (m *MetricsCollector) GetCurrentMetrics() map[string]interface{} {
	// Get the current values of the metrics
	requestsTotal := atomic.LoadUint64(&m.requestsTotal)
	requestsSucceeded := atomic.LoadUint64(&m.requestsSucceeded)
	requestsFailed := atomic.LoadUint64(&m.requestsFailed)
	currentConcurrent := atomic.LoadInt64(&m.currentConcurrent)
	memoryUsage := atomic.LoadUint64(&m.memoryUsage)
	
	m.mutex.RLock()
	cpuUsage := m.cpuUsage
	m.mutex.RUnlock()
	
	// Calculate derived metrics
	uptime := time.Since(m.startTime)
	requestsPerSecond := float64(requestsTotal) / uptime.Seconds()
	
	// Calculate response time percentiles
	p50 := m.responseTimes.GetPercentile(50)
	p90 := m.responseTimes.GetPercentile(90)
	p99 := m.responseTimes.GetPercentile(99)
	avgResponseTime := m.responseTimes.Average()
	
	// Calculate success rate
	var successRate float64
	if requestsTotal > 0 {
		successRate = float64(requestsSucceeded) / float64(requestsTotal) * 100.0
	}
	
	// Calculate server load as a ratio of current concurrent requests to maximum
	serverLoad := float64(currentConcurrent) / float64(m.maxConcurrent)
	
	// Return the metrics as a map
	return map[string]interface{}{
		"uptime":              uptime.String(),
		"requests_total":      requestsTotal,
		"requests_succeeded":  requestsSucceeded,
		"requests_failed":     requestsFailed,
		"requests_per_second": fmt.Sprintf("%.2f", requestsPerSecond),
		"success_rate":        fmt.Sprintf("%.2f%%", successRate),
		"concurrent_requests": currentConcurrent,
		"max_concurrent":      m.maxConcurrent,
		"server_load":         fmt.Sprintf("%.2f/10", serverLoad*10),
		"memory_usage":        fmt.Sprintf("%.2f MB", float64(memoryUsage)/1024/1024),
		"cpu_usage":           fmt.Sprintf("%.2f%%", cpuUsage*100),
		"p50_response_time":   p50.String(),
		"p90_response_time":   p90.String(),
		"p99_response_time":   p99.String(),
		"avg_response_time":   avgResponseTime.String(),
	}
}

// GetStatsReport returns a formatted string with the server statistics
func (m *MetricsCollector) GetStatsReport() string {
	metrics := m.GetCurrentMetrics()
	
	return fmt.Sprintf(`## Web server statistics
### uptime - %s
### requests_total - %d
### requests_succeeded - %d
### requests_failed - %d
### requests_per_second - %s
### success_rate - %s
### concurrent_requests - %d
### max_concurrent - %d
### server_load - %s
### memory_usage - %s
### cpu_usage - %s
### p50_response_time - %s
### p90_response_time - %s
### p99_response_time - %s
### avg_response_time - %s`,
		metrics["uptime"],
		metrics["requests_total"],
		metrics["requests_succeeded"],
		metrics["requests_failed"],
		metrics["requests_per_second"],
		metrics["success_rate"],
		metrics["concurrent_requests"],
		metrics["max_concurrent"],
		metrics["server_load"],
		metrics["memory_usage"],
		metrics["cpu_usage"],
		metrics["p50_response_time"],
		metrics["p90_response_time"],
		metrics["p99_response_time"],
		metrics["avg_response_time"])
}

// Shutdown stops the metrics collector
func (m *MetricsCollector) Shutdown() {
	close(m.stopCh)
}

// GetRequestTotal returns the total number of requests
func (m *MetricsCollector) GetRequestTotal() uint64 {
	return atomic.LoadUint64(&m.requestsTotal)
}

// GetRequestSucceeded returns the number of successful requests
func (m *MetricsCollector) GetRequestSucceeded() uint64 {
	return atomic.LoadUint64(&m.requestsSucceeded)
}

// GetRequestFailed returns the number of failed requests
func (m *MetricsCollector) GetRequestFailed() uint64 {
	return atomic.LoadUint64(&m.requestsFailed)
}

// GetCurrentConcurrent returns the current number of concurrent requests
func (m *MetricsCollector) GetCurrentConcurrent() int64 {
	return atomic.LoadInt64(&m.currentConcurrent)
}

// GetMaxConcurrent returns the maximum number of concurrent requests
func (m *MetricsCollector) GetMaxConcurrent() int64 {
	return m.maxConcurrent
}

// GetMemoryUsage returns the current memory usage in bytes
func (m *MetricsCollector) GetMemoryUsage() uint64 {
	return atomic.LoadUint64(&m.memoryUsage)
}

// GetCPUUsage returns the current CPU usage (0-1)
func (m *MetricsCollector) GetCPUUsage() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.cpuUsage
}

// GetResponseTimePercentile returns the nth percentile of response times
func (m *MetricsCollector) GetResponseTimePercentile(percentile float64) time.Duration {
	return m.responseTimes.GetPercentile(percentile)
}

// GetAverageResponseTime returns the average response time
func (m *MetricsCollector) GetAverageResponseTime() time.Duration {
	return m.responseTimes.Average()
}

// GetUptime returns the server uptime
func (m *MetricsCollector) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// Make the update methods public to allow direct updating
func (m *MetricsCollector) UpdateMemoryUsage() {
	m.updateMemoryUsage()
}

func (m *MetricsCollector) UpdateCPUUsage() {
	m.updateCPUUsage()
}
