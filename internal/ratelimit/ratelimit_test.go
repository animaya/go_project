package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	// Create a rate limiter with 10 tokens per second and capacity of 5 tokens
	limiter := NewTokenBucketLimiter(10, 5)
	
	// Test that we can take 5 tokens immediately
	for i := 0; i < 5; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected token %d to be allowed, but it was denied", i)
		}
	}
	
	// Test that the 6th token is denied
	if limiter.TryAllow() {
		t.Errorf("Expected 6th token to be denied, but it was allowed")
	}
	
	// Wait for a token to be replenished (should take about 100ms)
	time.Sleep(120 * time.Millisecond)
	
	// Test that we can now take one more token
	if !limiter.TryAllow() {
		t.Errorf("Expected token to be allowed after waiting, but it was denied")
	}
	
	// Test that the next token is denied again
	if limiter.TryAllow() {
		t.Errorf("Expected token to be denied, but it was allowed")
	}
}

func TestTokenBucketLimiterWithContext(t *testing.T) {
	// Create a rate limiter with 10 tokens per second and capacity of 1 token
	limiter := NewTokenBucketLimiter(10, 1)
	
	// Take the only token
	if !limiter.TryAllow() {
		t.Errorf("Expected token to be allowed, but it was denied")
	}
	
	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	// Try to get another token, which should be denied due to timeout
	if limiter.Allow(ctx) {
		t.Errorf("Expected token to be denied due to context timeout, but it was allowed")
	}
	
	// Wait for a token to be replenished
	time.Sleep(120 * time.Millisecond)
	
	// Create a new context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Try to get another token, which should be allowed now
	if !limiter.Allow(ctx) {
		t.Errorf("Expected token to be allowed after waiting, but it was denied")
	}
}

func TestSlidingWindowLimiter(t *testing.T) {
	// Create a sliding window limiter with 3 requests per 500ms
	limiter := NewSlidingWindowLimiter(3, 500*time.Millisecond)
	
	// Test that we can make 3 requests immediately
	for i := 0; i < 3; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected request %d to be allowed, but it was denied", i)
		}
	}
	
	// Test that the 4th request is denied
	if limiter.TryAllow() {
		t.Errorf("Expected 4th request to be denied, but it was allowed")
	}
	
	// Wait for the window to slide (should take about 500ms)
	time.Sleep(510 * time.Millisecond)
	
	// Test that we can now make 3 more requests
	for i := 0; i < 3; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected request %d to be allowed after waiting, but it was denied", i)
		}
	}
	
	// Test that the 4th request is denied again
	if limiter.TryAllow() {
		t.Errorf("Expected 4th request to be denied after waiting, but it was allowed")
	}
}

func TestSlidingWindowLimiterWithContext(t *testing.T) {
	// Create a sliding window limiter with 1 request per 500ms
	limiter := NewSlidingWindowLimiter(1, 500*time.Millisecond)
	
	// Take the only token
	if !limiter.TryAllow() {
		t.Errorf("Expected request to be allowed, but it was denied")
	}
	
	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	// Try to make another request, which should be denied due to timeout
	if limiter.Allow(ctx) {
		t.Errorf("Expected request to be denied due to context timeout, but it was allowed")
	}
	
	// Wait for the window to slide
	time.Sleep(510 * time.Millisecond)
	
	// Create a new context
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Try to make another request, which should be allowed now
	if !limiter.Allow(ctx) {
		t.Errorf("Expected request to be allowed after waiting, but it was denied")
	}
}

func TestDistributedLimiter(t *testing.T) {
	// Create a local limiter with 100 requests per second
	localLimiter := NewTokenBucketLimiter(100, 100)
	
	// Create a distributed limiter with 5 global tokens and a factor of 0.1
	limiter := NewDistributedLimiter(localLimiter, 0.1, 5)
	
	// Test that we can make 5 requests immediately
	for i := 0; i < 5; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected request %d to be allowed, but it was denied", i)
		}
	}
	
	// Test that the 6th request is denied
	if limiter.TryAllow() {
		t.Errorf("Expected 6th request to be denied, but it was allowed")
	}
	
	// Wait for a token to be released
	time.Sleep(120 * time.Millisecond)
	
	// Test that we can now make another request
	if !limiter.TryAllow() {
		t.Errorf("Expected request to be allowed after waiting, but it was denied")
	}
}

func TestCompositeRateLimiter(t *testing.T) {
	// Create two rate limiters with different rates
	limiter1 := NewTokenBucketLimiter(10, 5)  // 10 tokens per second, capacity 5
	limiter2 := NewTokenBucketLimiter(20, 10) // 20 tokens per second, capacity 10
	
	// Create a composite limiter
	limiter := NewCompositeRateLimiter(limiter1, limiter2)
	
	// Test that we can take 5 tokens immediately (limited by limiter1)
	for i := 0; i < 5; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected token %d to be allowed, but it was denied", i)
		}
	}
	
	// Test that the 6th token is denied (due to limiter1)
	if limiter.TryAllow() {
		t.Errorf("Expected 6th token to be denied, but it was allowed")
	}
	
	// Wait for a token to be replenished in limiter1
	time.Sleep(120 * time.Millisecond)
	
	// Test that we can now take one more token
	if !limiter.TryAllow() {
		t.Errorf("Expected token to be allowed after waiting, but it was denied")
	}
	
	// Drain limiter2 to test its limits
	// Limiter1 has 0 tokens left, limiter2 has 9 tokens left
	// We've already taken 6 tokens total
	
	// First, restore limiter1 to full capacity
	time.Sleep(500 * time.Millisecond) // Wait for 5 tokens to be replenished
	
	// Now take all the tokens from limiter2
	for i := 0; i < 10; i++ {
		limiter2.TryAllow() // Just drain the tokens
	}
	
	// Test that the next request is denied (due to limiter2)
	if limiter.TryAllow() {
		t.Errorf("Expected token to be denied due to limiter2, but it was allowed")
	}
}

func TestAdaptiveRateLimiter(t *testing.T) {
	// Create a base limiter
	baseLimiter := NewTokenBucketLimiter(10, 5)
	
	// Create an adaptive limiter
	limiter := NewAdaptiveRateLimiter(baseLimiter, 1, 20)
	defer limiter.Shutdown()
	
	// Test that the limiter works initially
	for i := 0; i < 5; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected token %d to be allowed, but it was denied", i)
		}
	}
	
	// Test that the 6th token is denied
	if limiter.TryAllow() {
		t.Errorf("Expected 6th token to be denied, but it was allowed")
	}
	
	// Wait for tokens to be replenished
	time.Sleep(600 * time.Millisecond) // Should get 6 tokens back (10 per second)
	
	// Test that we can take 5 more tokens
	for i := 0; i < 5; i++ {
		if !limiter.TryAllow() {
			t.Errorf("Expected token %d to be allowed after waiting, but it was denied", i)
		}
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	// Create a rate limiter with 100 tokens per second and capacity of 50 tokens
	limiter := NewTokenBucketLimiter(100, 50)
	
	// Number of concurrent goroutines
	numGoroutines := 100
	
	// Number of requests per goroutine
	requestsPerGoroutine := 10
	
	// Track the number of allowed and denied requests
	var allowed, denied int64
	
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	// Launch goroutines to make requests concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			
			// Make requests
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.TryAllow() {
					atomic.AddInt64(&allowed, 1)
				} else {
					atomic.AddInt64(&denied, 1)
				}
				
				// Add a small sleep to simulate work
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Log the results
	t.Logf("Allowed: %d, Denied: %d", allowed, denied)
	
	// Check that the total number of requests is correct
	totalRequests := numGoroutines * requestsPerGoroutine
	if int(allowed+denied) != totalRequests {
		t.Errorf("Expected %d total requests, got %d", totalRequests, allowed+denied)
	}
	
	// Check that we allowed approximately the expected number of requests
	// We start with 50 tokens and generate 100 per second
	// The test should run for about 100ms, so we should generate about 10 more tokens
	// So we expect about 60 allowed requests
	// But this can vary, so we just check that it's reasonable
	if allowed < 50 || allowed > 70 {
		t.Errorf("Expected about 60 allowed requests, got %d", allowed)
	}
}
