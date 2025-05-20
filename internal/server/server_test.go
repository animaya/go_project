package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	// Create a server with default options
	options := DefaultServerOptions()
	server := NewServer(options)
	
	// Check that the server was created successfully
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}
	
	// Check that the server has the correct options
	if server.options.MaxConcurrentRequests != options.MaxConcurrentRequests {
		t.Errorf("Expected MaxConcurrentRequests to be %d, got %d", options.MaxConcurrentRequests, server.options.MaxConcurrentRequests)
	}
	
	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Error shutting down server: %v", err)
	}
}

func TestHandleGenerateNames(t *testing.T) {
	// Create a server with default options
	server := NewServer(DefaultServerOptions())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	
	// Create a test request with a valid payload
	payload := RequestPayload{
		SessionID:    "test-session",
		Letter:       "A",
		NumOfEntries: 5,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error marshaling payload: %v", err)
	}
	
	req, err := http.NewRequest("POST", "/generate", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	server.handleGenerateNames(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check the content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
	
	// Parse the response
	var response ResponsePayload
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}
	
	// Check the response
	if response.SessionID != payload.SessionID {
		t.Errorf("Handler returned wrong session ID: got %v want %v", response.SessionID, payload.SessionID)
	}
	
	if response.NumOfEntries != payload.NumOfEntries {
		t.Errorf("Handler returned wrong number of entries: got %v want %v", response.NumOfEntries, payload.NumOfEntries)
	}
	
	if len(response.Names) != payload.NumOfEntries {
		t.Errorf("Handler returned wrong number of names: got %v want %v", len(response.Names), payload.NumOfEntries)
	}
	
	// Check that all names start with the requested letter
	for i, name := range response.Names {
		if name[0] != payload.Letter[0] {
			t.Errorf("Name %d (%s) does not start with %s", i, name, payload.Letter)
		}
	}
	
	// Test with invalid method
	req, err = http.NewRequest("GET", "/generate", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	
	rr = httptest.NewRecorder()
	server.handleGenerateNames(rr, req)
	
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleStats(t *testing.T) {
	// Create a server with default options
	server := NewServer(DefaultServerOptions())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	
	// Create a test request
	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler
	server.handleStats(rr, req)
	
	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check the content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "text/plain")
	}
	
	// Check that the response contains some stats
	if len(rr.Body.String()) == 0 {
		t.Error("Handler returned empty stats")
	}
}

func TestIntegration(t *testing.T) {
	// Create a server with default options
	server := NewServer(DefaultServerOptions())
	
	// Create a test server
	ts := httptest.NewServer(server.createRouter())
	defer ts.Close()
	
	// Shutdown the server when the test is done
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	
	// Make a request to generate names
	payload := RequestPayload{
		SessionID:    "test-session",
		Letter:       "B",
		NumOfEntries: 10,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error marshaling payload: %v", err)
	}
	
	resp, err := http.Post(ts.URL+"/generate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
	
	// Parse the response
	var response ResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}
	
	// Check the response
	if response.SessionID != payload.SessionID {
		t.Errorf("Expected session ID %s, got %s", payload.SessionID, response.SessionID)
	}
	
	if response.NumOfEntries != payload.NumOfEntries {
		t.Errorf("Expected %d entries, got %d", payload.NumOfEntries, response.NumOfEntries)
	}
	
	// Test the stats endpoint
	resp, err = http.Get(ts.URL + "/stats")
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
	
	// Test rate limiting
	// Make many requests quickly to trigger rate limiting
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				http.Post(ts.URL+"/generate", "application/json", bytes.NewBuffer(payloadBytes))
			}
		}()
	}
	
	// Wait for the rate limiter to kick in
	time.Sleep(1 * time.Second)
	
	// Check that we can still make a request after waiting
	resp, err = http.Post(ts.URL+"/generate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	
	// The status code might be 200 or 429 depending on the rate limiter, but the server should not crash
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status OK or TooManyRequests, got %v", resp.Status)
	}
}
