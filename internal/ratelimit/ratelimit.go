package ratelimit

import (
	"context"
	"sync"
	"time"
)

// RateLimiter is an interface for rate limiting operations
type RateLimiter interface {
	// Allow checks if a request is allowed and blocks if necessary
	// Returns true if the request is allowed, false if the context is canceled
	Allow(ctx context.Context) bool

	// TryAllow checks if a request is allowed without blocking
	// Returns true if the request is allowed, false otherwise
	TryAllow() bool
}

// TokenBucketLimiter implements a token bucket rate limiter
type TokenBucketLimiter struct {
	rate           float64 // tokens per second
	capacity       int64   // maximum number of tokens
	tokens         int64   // current number of tokens
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(rate float64, capacity int64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:           rate,
		capacity:       capacity,
		tokens:         capacity, // Start with a full bucket
		lastRefillTime: time.Now(),
	}
}

// refill adds tokens to the bucket based on the elapsed time
func (l *TokenBucketLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefillTime).Seconds()
	l.lastRefillTime = now

	// Calculate the number of tokens to add
	newTokens := int64(elapsed * l.rate)
	if newTokens > 0 {
		l.tokens = min(l.capacity, l.tokens+newTokens)
	}
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// Allow checks if a request is allowed and blocks if necessary
func (l *TokenBucketLimiter) Allow(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			// Context canceled
			return false
		default:
			// Check if a token is available
			if l.TryAllow() {
				return true
			}

			// No token available, wait a bit and try again
			// Calculate time until next token
			waitTime := time.Duration(1000/l.rate) * time.Millisecond

			// Wait for the next token or context cancellation
			select {
			case <-ctx.Done():
				return false
			case <-time.After(waitTime):
				// Try again
			}
		}
	}
}

// TryAllow checks if a request is allowed without blocking
func (l *TokenBucketLimiter) TryAllow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Refill tokens based on elapsed time
	l.refill()

	// Check if a token is available
	if l.tokens > 0 {
		l.tokens--
		return true
	}

	return false
}

// SlidingWindowLimiter implements a sliding window rate limiter
type SlidingWindowLimiter struct {
	maxRequests    int64         // maximum number of requests per window
	windowDuration time.Duration // duration of the window
	mutex          sync.Mutex
	requests       []time.Time // timestamps of recent requests
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(maxRequests int64, windowDuration time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		maxRequests:    maxRequests,
		windowDuration: windowDuration,
		requests:       make([]time.Time, 0, maxRequests),
	}
}

// pruneExpiredRequests removes expired requests from the window
func (l *SlidingWindowLimiter) pruneExpiredRequests() {
	now := time.Now()
	cutoff := now.Add(-l.windowDuration)

	// Find the index of the first non-expired request
	i := 0
	for i < len(l.requests) && l.requests[i].Before(cutoff) {
		i++
	}

	// Remove expired requests
	if i > 0 {
		l.requests = l.requests[i:]
	}
}

// Allow checks if a request is allowed and blocks if necessary
func (l *SlidingWindowLimiter) Allow(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			// Context canceled
			return false
		default:
			// Try to acquire a token
			if l.TryAllow() {
				return true
			}

			// No token available, calculate the wait time
			l.mutex.Lock()
			waitTime := l.windowDuration
			if len(l.requests) > 0 {
				// Wait until the oldest request expires
				expireTime := l.requests[0].Add(l.windowDuration)
				waitTime = time.Until(expireTime)
			}
			l.mutex.Unlock()

			// Wait for the next available slot or context cancellation
			select {
			case <-ctx.Done():
				return false
			case <-time.After(waitTime + time.Millisecond):
				// Try again
			}
		}
	}
}

// TryAllow checks if a request is allowed without blocking
func (l *SlidingWindowLimiter) TryAllow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Remove expired requests
	l.pruneExpiredRequests()

	// Check if we can add a new request
	if int64(len(l.requests)) < l.maxRequests {
		// Add the current request
		l.requests = append(l.requests, time.Now())
		return true
	}

	return false
}

// DistributedLimiter is a rate limiter that can be shared across multiple instances
// This is a simple implementation; in a real system, you would use a shared store
// like Redis to coordinate rate limiting across multiple servers
type DistributedLimiter struct {
	local         RateLimiter // Local rate limiter
	globalFactor  float64     // Factor to reduce local limit (0..1)
	sharedChannel chan struct{}
	mu            sync.Mutex
}

// NewDistributedLimiter creates a new distributed rate limiter
func NewDistributedLimiter(local RateLimiter, globalFactor float64, sharedLimit int) *DistributedLimiter {
	return &DistributedLimiter{
		local:         local,
		globalFactor:  globalFactor,
		sharedChannel: make(chan struct{}, sharedLimit),
	}
}

// Allow checks if a request is allowed and blocks if necessary
func (l *DistributedLimiter) Allow(ctx context.Context) bool {
	// First, check the local limiter
	if !l.local.Allow(ctx) {
		return false
	}

	// Then, check the global limiter
	select {
	case l.sharedChannel <- struct{}{}:
		// Acquired a global token
		// Release the token after some time (simulating request processing)
		go func() {
			// Release after some time (proportional to the global factor)
			releaseTime := time.Duration(float64(time.Second) * l.globalFactor)
			time.Sleep(releaseTime)
			<-l.sharedChannel
		}()
		return true
	case <-ctx.Done():
		// Context canceled
		return false
	}
}

// TryAllow checks if a request is allowed without blocking
func (l *DistributedLimiter) TryAllow() bool {
	// First, check the local limiter
	if !l.local.TryAllow() {
		return false
	}

	// Then, check the global limiter
	select {
	case l.sharedChannel <- struct{}{}:
		// Acquired a global token
		// Release the token after some time (simulating request processing)
		go func() {
			// Release after some time (proportional to the global factor)
			releaseTime := time.Duration(float64(time.Second) * l.globalFactor)
			time.Sleep(releaseTime)
			<-l.sharedChannel
		}()
		return true
	default:
		// No global token available
		return false
	}
}

// CompositeRateLimiter combines multiple rate limiters together
type CompositeRateLimiter struct {
	limiters []RateLimiter
}

// NewCompositeRateLimiter creates a new composite rate limiter
func NewCompositeRateLimiter(limiters ...RateLimiter) *CompositeRateLimiter {
	return &CompositeRateLimiter{
		limiters: limiters,
	}
}

// Allow checks if a request is allowed and blocks if necessary
// All limiters must allow the request for it to be allowed
func (l *CompositeRateLimiter) Allow(ctx context.Context) bool {
	// Check each limiter in sequence
	for _, limiter := range l.limiters {
		if !limiter.Allow(ctx) {
			return false
		}
	}

	return true
}

// TryAllow checks if a request is allowed without blocking
// All limiters must allow the request for it to be allowed
func (l *CompositeRateLimiter) TryAllow() bool {
	// Check each limiter in sequence
	for _, limiter := range l.limiters {
		if !limiter.TryAllow() {
			return false
		}
	}

	return true
}

// AdaptiveRateLimiter dynamically adjusts its rate limit based on system load
type AdaptiveRateLimiter struct {
	baseLimiter    RateLimiter
	minRate        float64
	maxRate        float64
	currentRate    float64
	loadThreshold  float64 // Value between 0 and 1, representing system load
	adjustInterval time.Duration
	mu             sync.Mutex
	stopCh         chan struct{}
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(baseLimiter RateLimiter, minRate, maxRate float64) *AdaptiveRateLimiter {
	limiter := &AdaptiveRateLimiter{
		baseLimiter:    baseLimiter,
		minRate:        minRate,
		maxRate:        maxRate,
		currentRate:    maxRate, // Start with max rate
		loadThreshold:  0.7,     // Default load threshold
		adjustInterval: 5 * time.Second,
		stopCh:         make(chan struct{}),
	}

	// Start the adjustment loop
	go limiter.adjustLoop()

	return limiter
}

// adjustLoop periodically adjusts the rate limit based on system load
func (l *AdaptiveRateLimiter) adjustLoop() {
	ticker := time.NewTicker(l.adjustInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Adjust the rate based on system load
			l.adjustRate()
		case <-l.stopCh:
			// Stop the adjustment loop
			return
		}
	}
}

// adjustRate adjusts the rate based on system load
func (l *AdaptiveRateLimiter) adjustRate() {
	// Get the current system load
	load := getSystemLoad()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Adjust the rate based on load
	if load > l.loadThreshold {
		// System is under heavy load, decrease the rate
		l.currentRate = maxFloat(l.minRate, l.currentRate*0.9)
	} else {
		// System is under normal load, increase the rate
		l.currentRate = minFloat(l.maxRate, l.currentRate*1.1)
	}
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// maxFloat returns the maximum of two float64 values
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// getSystemLoad returns a simulated system load between 0 and 1
// In a real system, this would be based on actual metrics (CPU, memory, etc.)
func getSystemLoad() float64 {
	// Simulate some load based on time of day
	now := time.Now()
	hour := now.Hour()

	// Higher load during business hours
	if hour >= 9 && hour <= 17 {
		return 0.5 + 0.3*float64(now.Minute())/60.0
	}

	return 0.3 + 0.2*float64(now.Minute())/60.0
}

// Allow checks if a request is allowed and blocks if necessary
func (l *AdaptiveRateLimiter) Allow(ctx context.Context) bool {
	return l.baseLimiter.Allow(ctx)
}

// TryAllow checks if a request is allowed without blocking
func (l *AdaptiveRateLimiter) TryAllow() bool {
	return l.baseLimiter.TryAllow()
}

// Shutdown stops the adaptive rate limiter's adjustment loop
func (l *AdaptiveRateLimiter) Shutdown() {
	close(l.stopCh)
}
