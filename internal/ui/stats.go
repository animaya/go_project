package ui

import (
	"html/template"
	"log"
	"strings"
)

// StatsTemplate holds the HTML template for statistics page
var StatsTemplate *template.Template

// Initialize initializes the UI templates
func Initialize() {
	var err error

	// Define our HTML template with HTMX integration
	const statsTemplateText = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Statistics</title>
    <script src="https://unpkg.com/htmx.org@1.9.11"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f9f9f9;
            color: #333;
        }
        .stats-dashboard {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .stat-card {
            background-color: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 6px 12px rgba(0, 0, 0, 0.15);
        }
        .stat-name {
            font-size: 1rem;
            color: #666;
            margin-bottom: 5px;
        }
        .stat-value {
            font-size: 1.75rem;
            font-weight: 600;
            margin-bottom: 0;
        }
        .server-state {
            margin-bottom: 20px;
            padding: 15px;
            border-radius: 8px;
            text-align: center;
            font-weight: bold;
        }
        .server-online {
            background-color: #d4edda;
            color: #155724;
        }
        .server-offline {
            background-color: #f8d7da;
            color: #721c24;
        }
        .stat-group {
            margin-bottom: 10px;
            font-size: 1.2rem;
            font-weight: bold;
            color: #2c3e50;
        }
        header {
            margin-bottom: 30px;
            border-bottom: 1px solid #eee;
            padding-bottom: 10px;
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 5px;
        }
        .subtitle {
            color: #7f8c8d;
            font-style: italic;
        }
        .response-times {
            grid-column: 1 / -1;
        }
        .response-card {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #4299e1;
        }
    </style>
</head>
<body>
    <header>
        <h1>Real-time Server Statistics</h1>
        <p class="subtitle">Name Generator Web Server Status Dashboard</p>
    </header>

    <!-- Server state indicator -->
    <div class="server-state server-online" id="server-state">
        Server Status: ONLINE
    </div>

    <!-- Stats container that will be refreshed via HTMX -->
    <div id="stats-container" hx-get="/stats/data" hx-trigger="load, every 2s" hx-swap="innerHTML">
        <!-- Initial loading state -->
        <p>Loading statistics...</p>
    </div>

    <script>
        // Function to check server status and update the indicator
        function updateServerStatus() {
            fetch('/stats/data')
                .then(response => {
                    // If we get a response, server is online
                    const statusElement = document.getElementById('server-state');
                    if (response.ok) {
                        statusElement.className = 'server-state server-online';
                        statusElement.textContent = 'Server Status: ONLINE';
                    } else {
                        throw new Error('Server returned an error');
                    }
                })
                .catch(error => {
                    // If there's an error, server is offline
                    const statusElement = document.getElementById('server-state');
                    statusElement.className = 'server-state server-offline';
                    statusElement.textContent = 'Server Status: OFFLINE';
                });
        }

        // Update server status initially and every 3 seconds
        updateServerStatus();
        setInterval(updateServerStatus, 3000);
    </script>
</body>
</html>
`

	const statsDataTemplateText = `
<div class="stats-dashboard">
    <!-- Server overview -->
    <div class="stat-card">
        <div class="stat-group">Server Overview</div>
        <div class="stat-name">Uptime</div>
        <div class="stat-value">{{.uptime}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Memory & CPU</div>
        <div class="stat-name">Memory Usage</div>
        <div class="stat-value">{{.memory_usage}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Memory & CPU</div>
        <div class="stat-name">CPU Usage</div>
        <div class="stat-value">{{.cpu_usage}}</div>
    </div>
    
    <!-- Request statistics -->
    <div class="stat-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Total Requests</div>
        <div class="stat-value">{{.requests_total}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Successful Requests</div>
        <div class="stat-value">{{.requests_succeeded}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Failed Requests</div>
        <div class="stat-value">{{.requests_failed}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Requests Per Second</div>
        <div class="stat-value">{{.requests_per_second}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Success Rate</div>
        <div class="stat-value">{{.success_rate}}</div>
    </div>
    
    <!-- Capacity information -->
    <div class="stat-card">
        <div class="stat-group">Server Capacity</div>
        <div class="stat-name">Current Concurrent Requests</div>
        <div class="stat-value">{{.concurrent_requests}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Server Capacity</div>
        <div class="stat-name">Maximum Concurrent</div>
        <div class="stat-value">{{.max_concurrent}}</div>
    </div>
    
    <div class="stat-card">
        <div class="stat-group">Server Capacity</div>
        <div class="stat-name">Server Load</div>
        <div class="stat-value">{{.server_load}}</div>
    </div>
    
    <!-- Response time metrics in a wider card -->
    <div class="stat-card response-times">
        <div class="stat-group">Response Time Metrics</div>
        <div class="response-card">
            <div class="stat-name">50th Percentile (P50)</div>
            <div class="stat-value">{{.p50_response_time}}</div>
        </div>
        <div class="response-card" style="margin-top: 10px;">
            <div class="stat-name">90th Percentile (P90)</div>
            <div class="stat-value">{{.p90_response_time}}</div>
        </div>
        <div class="response-card" style="margin-top: 10px;">
            <div class="stat-name">99th Percentile (P99)</div>
            <div class="stat-value">{{.p99_response_time}}</div>
        </div>
    </div>
</div>
`

	// Combine templates
	templateText := statsTemplateText
	
	// Parse the template
	StatsTemplate, err = template.New("stats").Parse(templateText)
	if err != nil {
		log.Fatalf("Failed to parse stats template: %v", err)
	}

	// Add the data template
	StatsTemplate, err = StatsTemplate.New("statsData").Parse(statsDataTemplateText)
	if err != nil {
		log.Fatalf("Failed to parse stats data template: %v", err)
	}
}

// ParseStatsReport converts a stats report string to a map for the template
func ParseStatsReport(report string) map[string]string {
	// Create a map to hold stats
	stats := make(map[string]string)
	
	// Split the report into lines
	lines := strings.Split(report, "\n")
	
	// Skip the first line (title)
	for _, line := range lines[1:] {
		// Check if the line contains a statistic
		if strings.Contains(line, " - ") {
			// Split the line into key and value
			parts := strings.SplitN(line, " - ", 2)
			if len(parts) == 2 {
				// Extract the key without the ### prefix
				key := strings.TrimSpace(strings.TrimPrefix(parts[0], "###"))
				value := strings.TrimSpace(parts[1])
				stats[key] = value
			}
		}
	}
	
	// Set default status
	stats["status"] = "ONLINE"
	
	return stats
}
