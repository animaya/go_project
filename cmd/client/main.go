package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
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
		return
	}
	
	// Create request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		atomic.AddUint64(&stats.FailedRequests, 1)
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
		return
	}
	defer resp.Body.Close()
	
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
	fmt.Println("================================================")
}

func main() {
	// Define command line flags
	serverURL := flag.String("url", "http://localhost:8080/generate", "Server URL")
	numClients := flag.Int("clients", 100, "Number of concurrent clients")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	flag.Parse()
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Initialize statistics
	stats := &ClientStats{
		MinLatency: 0,
		MaxLatency: 0,
	}
	
	// Print welcome message
	fmt.Printf("Starting client simulator with %d concurrent clients for %s\n", *numClients, *duration)
	fmt.Printf("Target server: %s\n", *serverURL)
	fmt.Println("Press Ctrl+C to stop the test early")
	
	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Start the timer
	startTime := time.Now()
	
	// Start the test
	stopTest := make(chan struct{})
	
	// Start client goroutines
	for i := 0; i < *numClients; i++ {
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
	
	// Print stats every 5 seconds during the test
	ticker := time.NewTicker(5 * time.Second)
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
	
	// Wait for the test duration
	time.Sleep(*duration)
	
	// Stop all client goroutines
	close(stopTest)
	
	// Stop the ticker
	ticker.Stop()
	
	// Wait for all requests to finish
	wg.Wait()
	
	// Calculate the actual test duration
	actualDuration := time.Since(startTime)
	
	// Print final statistics
	fmt.Println("\nTest completed!")
	printStats(stats, actualDuration)
}
