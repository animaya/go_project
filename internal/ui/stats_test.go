package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseStatsReport(t *testing.T) {
	// Create a sample stats report
	statsReport := `## Web server statistics
### uptime - 32.069411s
### requests_total - 6
### requests_succeeded - 5
### requests_failed - 0
### requests_per_second - 0.19`

	// Parse the stats report
	stats := ParseStatsReport(statsReport)

	// Check the parsed stats
	expected := map[string]string{
		"uptime":             "32.069411s",
		"requests_total":     "6",
		"requests_succeeded": "5",
		"requests_failed":    "0",
		"requests_per_second": "0.19",
		"status":             "ONLINE",
	}

	// Compare the parsed stats with the expected stats
	for key, expectedValue := range expected {
		value, ok := stats[key]
		if !ok {
			t.Errorf("Expected key %s not found in parsed stats", key)
		} else if value != expectedValue {
			t.Errorf("Expected value %s for key %s, got %s", expectedValue, key, value)
		}
	}
}

func TestTemplateRendering(t *testing.T) {
	// Initialize the UI templates
	Initialize()

	// Check if the template was created
	if StatsTemplate == nil {
		t.Fatal("StatsTemplate is nil after initialization")
	}

	// Create a sample data map for the template
	data := map[string]string{
		"uptime":              "32.069411s",
		"requests_total":      "6",
		"requests_succeeded":  "5",
		"requests_failed":     "0",
		"requests_per_second": "0.19",
		"success_rate":        "83.33%",
		"concurrent_requests": "1",
		"max_concurrent":      "1000",
		"server_load":         "0.01/10",
		"memory_usage":        "0.35 MB",
		"cpu_usage":           "0.72%",
		"p50_response_time":   "110.083µs",
		"p90_response_time":   "318.875µs",
		"p99_response_time":   "318.875µs",
	}

	// Try to render the template
	var buf bytes.Buffer
	err := StatsTemplate.ExecuteTemplate(&buf, "statsData", data)
	if err != nil {
		t.Fatalf("Failed to render statsData template: %v", err)
	}

	// Check if the rendered template contains expected data
	rendered := buf.String()
	for _, value := range data {
		if !strings.Contains(rendered, value) {
			t.Errorf("Rendered template does not contain expected value %s", value)
		}
	}

	// Clear the buffer and try rendering the main template
	buf.Reset()
	err = StatsTemplate.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to render main template: %v", err)
	}

	// Check if the rendered template contains expected elements
	rendered = buf.String()
	for _, expected := range []string{
		"Real-time Server Statistics",
		"Name Generator Web Server Status Dashboard",
		"Server Status: ONLINE",
		"Server Overview",
		"Request Statistics",
		"Server Capacity",
		"Response Time Metrics",
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("Rendered template does not contain expected element %s", expected)
		}
	}
}
