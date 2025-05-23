package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/amirahmetzanov/go_project/internal/cache"
	"github.com/amirahmetzanov/go_project/internal/generator"
	"github.com/amirahmetzanov/go_project/internal/metrics"
	"github.com/amirahmetzanov/go_project/internal/ratelimit"
	"github.com/amirahmetzanov/go_project/internal/ui"
)

// RequestPayload represents the JSON payload in the incoming request
type RequestPayload struct {
	SessionID     string `json:"session_id"`
	Letter        string `json:"letter"`
	NumOfEntries  int    `json:"num_of_entries"`
}

// ResponsePayload represents the JSON response sent back to the client
type ResponsePayload struct {
	SessionID     string   `json:"session_id"`
	Names         []string `json:"names"`
	NumOfEntries  int      `json:"num_of_entries"`
}

// ServerOptions represents configuration options for the server
type ServerOptions struct {
	MaxConcurrentRequests int64
	RequestRateLimit      float64 // Requests per second
	CacheSize             int
	CacheExpiration       time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
}

// DefaultServerOptions returns the default server options
func DefaultServerOptions() ServerOptions {
	return ServerOptions{
		MaxConcurrentRequests: 5000,         // Significantly increased from 2000 to 5000
		RequestRateLimit:      2000,         // Doubled from 1000 to 2000 requests per second
		CacheSize:             5000,         // Significantly increased cache size for high concurrency
		CacheExpiration:       10 * time.Minute, // Doubled cache expiration to reduce computation
		ReadTimeout:           15 * time.Second, // Increased for very high concurrent load
		WriteTimeout:          20 * time.Second, // Increased for very high concurrent load
		IdleTimeout:           60 * time.Second,
	}
}

// responseWriter is a custom ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Server represents our web server instance
type Server struct {
	metrics        *metrics.MetricsCollector
	nameGenerator  *generator.NameGenerator
	cache          *cache.ConcurrentLRUCache
	rateLimiter    ratelimit.RateLimiter
	httpServer     *http.Server
	options        ServerOptions
}

// NewServer creates a new server instance with the given options
func NewServer(options ServerOptions) *Server {
	// Create a metrics collector
	metricsCollector := metrics.NewMetricsCollector(options.MaxConcurrentRequests)
	
	// Create a name generator with many more workers for extreme concurrency
	nameGenerator := generator.NewNameGenerator(16) // Increased from 8 to 16 workers
	
	// Create a cache with many more shards for extreme concurrency
	cacheInstance := cache.NewConcurrentLRUCache(
		options.CacheSize,
		64, // Significantly increased from 32 to 64 shards for extreme concurrency
		options.CacheExpiration,
		options.CacheExpiration/2, // Cleanup at half the expiration time
	)
	
	// Create a rate limiter
	// Use a token bucket rate limiter with 30x burst capacity - extreme burst capacity
	burstCapacity := int64(options.RequestRateLimit * 30)
	tokenLimiter := ratelimit.NewTokenBucketLimiter(options.RequestRateLimit, burstCapacity)
	
	// Create a sliding window rate limiter with much higher allowance
	slidingLimiter := ratelimit.NewSlidingWindowLimiter(
		int64(options.RequestRateLimit*2.0), // Allow double the requests in sliding window
		time.Second,
	)
	
	// Create a composite rate limiter that uses both strategies
	compositeLimiter := ratelimit.NewCompositeRateLimiter(tokenLimiter, slidingLimiter)
	
	// Create the server
	server := &Server{
		metrics:       metricsCollector,
		nameGenerator: nameGenerator,
		cache:         cacheInstance,
		rateLimiter:   compositeLimiter,
		options:       options,
	}
	
	// Get port from environment variable with fallback to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Create the HTTP server
	server.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      server.createRouter(),
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
		IdleTimeout:  options.IdleTimeout,
	}
	
	return server
}

// createRouter creates the HTTP router for the server
func (s *Server) createRouter() http.Handler {
	mux := http.NewServeMux()
	
	// Register the routes
	mux.HandleFunc("/generate", s.handleGenerateNames)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/stats/data", s.handleStats)
	
	// Create a middleware chain
	handler := s.metricsMiddleware(
		s.loggingMiddleware(
			s.rateLimitMiddleware(
				mux,
			),
		),
	)
	
	return handler
}

// metricsMiddleware tracks request metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the start of the request
		done := s.metrics.RecordRequest()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Record the end of the request
		done(nil)
	})
}

// loggingMiddleware logs request information
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the start time
		start := time.Now()
		
		// Create a custom response writer to capture the status code
		responseWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the next handler
		next.ServeHTTP(responseWriter, r)
		
		// Log the request
		log.Printf("[%s] %s %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			responseWriter.statusCode,
			time.Since(start),
		)
	})
}

// rateLimitMiddleware applies rate limiting to requests
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a context with a timeout - increased to 2 seconds
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		
		// Check the rate limiter
		if !s.rateLimiter.Allow(ctx) {
			// Return a more informative error message with retry-after header
			w.Header().Set("Retry-After", "1") // Suggest client to retry after 1 second
			http.Error(w, "Rate limit exceeded, please try again later", http.StatusTooManyRequests)
			
			// Log rate limiting events to help diagnose issues
			log.Printf("Rate limit exceeded for request from %s to %s", r.RemoteAddr, r.URL.Path)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// getCacheKey generates a cache key for the given request
func getCacheKey(letter string, count int) string {
	return fmt.Sprintf("%s:%d", letter, count)
}

// handleGenerateNames handles the name generation request
func (s *Server) handleGenerateNames(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var payload RequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request payload
	if payload.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}
	
	if payload.NumOfEntries <= 0 {
		payload.NumOfEntries = 1 // Default to 1 if not specified
	} else if payload.NumOfEntries > 100 {
		payload.NumOfEntries = 100 // Limit to 100 to prevent abuse
	}

	// Generate the cache key
	cacheKey := getCacheKey(payload.Letter, payload.NumOfEntries)

	// Try to get the names from the cache
	if cachedNames, found := s.cache.Get(cacheKey); found {
		// Found in cache, return the cached names
		response := ResponsePayload{
			SessionID:    payload.SessionID,
			Names:        cachedNames.([]string),
			NumOfEntries: len(cachedNames.([]string)),
		}

		// Set the content type header
		w.Header().Set("Content-Type", "application/json")

		// Encode the response
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Not found in cache, generate new names
	// Create a context with a timeout for name generation
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Generate names with the context
	names := s.nameGenerator.GenerateWithContext(ctx, payload.Letter, payload.NumOfEntries)

	// Cache the generated names
	s.cache.Set(cacheKey, names)

	// Prepare the response
	response := ResponsePayload{
		SessionID:    payload.SessionID,
		Names:        names,
		NumOfEntries: len(names),
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode the response
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleStats handles the statistics display request
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	// Force metrics update before responding
	s.metrics.UpdateMemoryUsage()
	s.metrics.UpdateCPUUsage()
	
	// Check if this is a request for the HTML page or for the stats data
	if r.URL.Path == "/stats/data" {
		// Return just the stats data for HTMX to update
		w.Header().Set("Content-Type", "text/html")
		
		// Get the stats data
		metrics := s.metrics.GetCurrentMetrics()
		
		// Set cache control headers to prevent caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		
		// Execute the template with the stats data
		if err := ui.StatsTemplate.ExecuteTemplate(w, "statsData", metrics); err != nil {
			http.Error(w, "Failed to render stats data", http.StatusInternalServerError)
			log.Printf("Error rendering stats data: %v", err)
		}
		return
	}
	
	// Return the full HTML page
	w.Header().Set("Content-Type", "text/html")
	
	// Set cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	// Execute the template with the stats data
	metrics := s.metrics.GetCurrentMetrics()
	if err := ui.StatsTemplate.Execute(w, metrics); err != nil {
		http.Error(w, "Failed to render stats page", http.StatusInternalServerError)
		log.Printf("Error rendering stats page: %v", err)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Initialize UI templates
	ui.Initialize()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")

	// Shutdown the HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	// Shutdown the metrics collector
	s.metrics.Shutdown()

	// Shutdown the name generator
	s.nameGenerator.Shutdown()

	// Shutdown the cache
	s.cache.Shutdown()

	log.Println("Server stopped")
	return nil
}
