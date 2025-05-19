package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/amirahmetzanov/go_project/internal/generator"
	"github.com/amirahmetzanov/go_project/internal/stats"
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

// Server represents our web server instance
type Server struct {
	stats *stats.ServerStats
	requestLimiter chan struct{} // Used to limit concurrent requests
	mu            sync.Mutex
}

// NewServer creates a new server instance
func NewServer(maxConcurrent int64) *Server {
	serverStats := stats.NewServerStats()
	serverStats.MaxConcurrent = maxConcurrent
	
	// Start periodic monitoring
	serverStats.StartMonitoring(5 * time.Second)
	
	return &Server{
		stats: serverStats,
		requestLimiter: make(chan struct{}, maxConcurrent),
	}
}

// handleGenerateNames handles the name generation request
func (s *Server) handleGenerateNames(w http.ResponseWriter, r *http.Request) {
	// Rate limiting: try to acquire a token
	select {
	case s.requestLimiter <- struct{}{}:
		// Token acquired, continue processing
		defer func() { <-s.requestLimiter }() // Release token when done
	default:
		// No tokens available, server is at capacity
		http.Error(w, "Server too busy, try again later", http.StatusServiceUnavailable)
		return
	}

	// Increment concurrent requests counter
	s.stats.IncConcurrent()
	defer s.stats.DecConcurrent()

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

	// Generate names
	names := generator.GenerateNames(payload.Letter, payload.NumOfEntries)

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

	// Increment the processed requests counter
	s.stats.IncrementRequests()
}

// handleStats handles the statistics display request
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, s.stats.GetStatsReport())
}

func main() {
	// Create a new server with maximum 1000 concurrent requests
	server := NewServer(1000)
	
	// Set up HTTP routes
	http.HandleFunc("/generate", server.handleGenerateNames)
	http.HandleFunc("/stats", server.handleStats)
	
	// Start HTTP server
	httpServer := &http.Server{
		Addr: ":8080",
	}
	
	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port 8080...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")
	
	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Error during server shutdown: %v", err)
	}
	
	log.Println("Server stopped")
}
