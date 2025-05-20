package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// RequestPayload represents the JSON payload in the request
type RequestPayload struct {
	SessionID    string `json:"session_id"`
	Letter       string `json:"letter"`
	NumOfEntries int    `json:"num_of_entries"`
}

// ResponsePayload represents the JSON response from the server
type ResponsePayload struct {
	SessionID    string   `json:"session_id"`
	Names        []string `json:"names"`
	NumOfEntries int      `json:"num_of_entries"`
}

// ClientStats tracks performance metrics
type ClientStats struct {
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	TotalLatency       uint64 // in milliseconds
	MaxLatency         uint64 // in milliseconds
	MinLatency         uint64 // in milliseconds
	StatusCodes        map[int]uint64
	Errors             map[string]uint64
	mutex              sync.RWMutex
}

// NewClientStats creates a new client stats instance
func NewClientStats() *ClientStats {
	return &ClientStats{
		StatusCodes: make(map[int]uint64),
		Errors:      make(map[string]uint64),
	}
}

// IncrementStatusCode increments the count for a specific status code
func (s *ClientStats) IncrementStatusCode(code int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.StatusCodes[code]++
}

// IncrementError increments the count for a specific error
func (s *ClientStats) IncrementError(err string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Errors[err]++
}

// generateRandomSessionID generates a random session ID
func generateRandomSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 10
	
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	
	return string(b)
}

// generateRandomLetter generates a random capital letter
func generateRandomLetter() string {
	return string(rune('A' + rand.Intn(26)))
}

// sendRequest sends a single request to the server
func sendRequest(serverURL string, stats *ClientStats, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// Generate random parameters
	sessionID := generateRandomSessionID()
	letter := generateRandomLetter()
	numOfEntries := rand.Intn(20) + 1 // Random number between 1 and 20
	
	// Create request payload
	payload := RequestPayload{
		SessionID:    sessionID,
		Letter:       letter,
		NumOfEntries: numOfEntries,
	}
	
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError(fmt.Sprintf("marshal: %v", err))
		return
	}
	
	// Create request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError(fmt.Sprintf("create: %v", err))
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Send request and measure time
	startTime := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	latency := time.Since(startTime).Milliseconds()
	
	// Update total requests counter
	atomic.AddUint64(&stats.TotalRequests, 1)
	
	// Update latency statistics
	atomic.AddUint64(&stats.TotalLatency, uint64(latency))
	
	// Update min latency (atomically)
	for {
		min := atomic.LoadUint64(&stats.MinLatency)
		if min == 0 || uint64(latency) < min {
			if atomic.CompareAndSwapUint64(&stats.MinLatency, min, uint64(latency)) {
				break
			}
		} else {
			break
		}
	}
	
	// Update max latency (atomically)
	for {
		max := atomic.LoadUint64(&stats.MaxLatency)
		if uint64(latency) > max {
			if atomic.CompareAndSwapUint64(&stats.MaxLatency, max, uint64(latency)) {
				break
			}
		} else {
			break
		}
	}
	
	// Check for errors
	if err != nil {
		log.Printf("Error sending request: %v", err)
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError(fmt.Sprintf("send: %v", err))
		return
	}
	defer resp.Body.Close()
	
	// Update status code counter
	stats.IncrementStatusCode(resp.StatusCode)
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response: %s", resp.Status)
		atomic.AddUint64(&stats.FailedRequests, 1)
		return
	}
	
	// Parse response
	var responsePayload ResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&responsePayload); err != nil {
		log.Printf("Error decoding response: %v", err)
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError(fmt.Sprintf("decode: %v", err))
		return
	}
	
	// Validate response
	if responsePayload.SessionID != sessionID {
		log.Printf("Session ID mismatch: expected %s, got %s", sessionID, responsePayload.SessionID)
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError("session_id_mismatch")
		return
	}
	
	if len(responsePayload.Names) != numOfEntries {
		log.Printf("Number of entries mismatch: expected %d, got %d", numOfEntries, len(responsePayload.Names))
		atomic.AddUint64(&stats.FailedRequests, 1)
		stats.IncrementError("num_entries_mismatch")
		return
	}
	
	// Request was successful
	atomic.AddUint64(&stats.SuccessfulRequests, 1)
}

// printStats prints the current statistics
func printStats(stats *ClientStats, duration time.Duration) {
	totalRequests := atomic.LoadUint64(&stats.TotalRequests)
	successfulRequests := atomic.LoadUint64(&stats.SuccessfulRequests)
	failedRequests := atomic.LoadUint64(&stats.FailedRequests)
	totalLatency := atomic.LoadUint64(&stats.TotalLatency)
	maxLatency := atomic.LoadUint64(&stats.MaxLatency)
	minLatency := atomic.LoadUint64(&stats.MinLatency)
	
	var avgLatency uint64
	if totalRequests > 0 {
		avgLatency = totalLatency / totalRequests
	}
	
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	
	fmt.Println("========== Client Simulator Statistics ==========")
	fmt.Printf("Total Requests:       %d\n", totalRequests)
	fmt.Printf("Successful Requests:  %d (%.2f%%)\n", successfulRequests, float64(successfulRequests)/float64(totalRequests)*100)
	fmt.Printf("Failed Requests:      %d (%.2f%%)\n", failedRequests, float64(failedRequests)/float64(totalRequests)*100)
	fmt.Printf("Requests Per Second:  %.2f\n", requestsPerSecond)
	fmt.Printf("Min Latency:          %d ms\n", minLatency)
	fmt.Printf("Avg Latency:          %d ms\n", avgLatency)
	fmt.Printf("Max Latency:          %d ms\n", maxLatency)
	
	// Print status code distribution
	fmt.Println("\nStatus Code Distribution:")
	stats.mutex.RLock()
	for code, count := range stats.StatusCodes {
		fmt.Printf("  %d: %d (%.2f%%)\n", code, count, float64(count)/float64(totalRequests)*100)
	}
	stats.mutex.RUnlock()
	
	// Print error distribution
	fmt.Println("\nError Distribution:")
	stats.mutex.RLock()
	if len(stats.Errors) == 0 {
		fmt.Println("  No errors")
	} else {
		for err, count := range stats.Errors {
			fmt.Printf("  %s: %d (%.2f%%)\n", err, count, float64(count)/float64(totalRequests)*100)
		}
	}
	stats.mutex.RUnlock()
	
	fmt.Println("================================================")
}

func main() {
	// Define command line flags
	serverURL := flag.String("url", "http://localhost:8080/generate", "Server URL")
	numClients := flag.Int("clients", 100, "Number of concurrent clients")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	rampUp := flag.Duration("ramp-up", 5*time.Second, "Ramp-up duration")
	statsInterval := flag.Duration("stats-interval", 5*time.Second, "Stats printing interval")
	flag.Parse()
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Initialize statistics
	stats := NewClientStats()
	
	// Print welcome message
	fmt.Printf("Starting client simulator with %d concurrent clients for %s\n", *numClients, *duration)
	fmt.Printf("Target server: %s\n", *serverURL)
	fmt.Printf("Ramp-up duration: %s\n", *rampUp)
	fmt.Println("Press Ctrl+C to stop the test early")
	
	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Start the timer
	startTime := time.Now()
	
	// Start the test
	stopTest := make(chan struct{})
	
	// Calculate ramp-up interval
	rampUpInterval := time.Duration(int64(*rampUp) / int64(*numClients))
	
	// Start client goroutines with ramp-up
	for i := 0; i < *numClients; i++ {
		// Add a delay for ramp-up
		if *rampUp > 0 {
			time.Sleep(rampUpInterval)
		}
		
		go func() {
			for {
				select {
				case <-stopTest:
					return
				default:
					wg.Add(1)
					sendRequest(*serverURL, stats, &wg)
					
					// Add some randomization to request timing
					sleepTime := time.Duration(rand.Intn(100)) * time.Millisecond
					time.Sleep(sleepTime)
				}
			}
		}()
	}
	
	// Print stats every interval during the test
	ticker := time.NewTicker(*statsInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				printStats(stats, time.Since(startTime))
			case <-stopTest:
				return
			}
		}
	}()
	
	// Setup signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	
	// Wait for test duration or interrupt
	select {
	case <-time.After(*duration):
		fmt.Println("Test duration reached, stopping...")
	case sig := <-signalCh:
		fmt.Printf("Received signal %v, stopping...\n", sig)
	}
	
	// Stop all client goroutines
	close(stopTest)
	
	// Stop the ticker
	ticker.Stop()
	
	// Wait for all requests to finish (with timeout)
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
	
	select {
	case <-waitCh:
		// All requests completed
	case <-time.After(5 * time.Second):
		fmt.Println("Timed out waiting for requests to complete")
	}
	
	// Calculate the actual test duration
	actualDuration := time.Since(startTime)
	
	// Print final statistics
	fmt.Println("\nTest completed!")
	printStats(stats, actualDuration)
	
	// Print server stats
	fmt.Println("\nFetching server statistics...")
	resp, err := http.Get(strings.TrimSuffix(*serverURL, "/generate") + "/stats")
	if err != nil {
		fmt.Printf("Error fetching server stats: %v\n", err)
	} else {
		defer resp.Body.Close()
		
		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading server stats: %v\n", err)
		} else {
			fmt.Println("\nServer Statistics:")
			fmt.Println(string(body))
		}
	}
}
