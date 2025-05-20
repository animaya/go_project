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
	// Define our HTML template with HTMX integration
	const statsHTML = `<!DOCTYPE html>
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
            background-color: #f0f2f5;
            color: #333;
        }
        .stats-dashboard {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .stat-card {
            background-color: white;
            border-radius: 12px;
            padding: 25px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            transition: all 0.3s ease;
            border-top: 4px solid #4361ee;
        }
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
        }
        .stat-name {
            font-size: 1.1rem;
            color: #666;
            margin-bottom: 8px;
            font-weight: 500;
        }
        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0;
            color: #2d3748;
        }
        .server-state {
            margin-bottom: 25px;
            padding: 18px;
            border-radius: 12px;
            text-align: center;
            font-weight: bold;
            font-size: 1.2rem;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }
        .server-online {
            background-color: #d4edda;
            color: #155724;
            border-left: 5px solid #28a745;
        }
        .server-offline {
            background-color: #f8d7da;
            color: #721c24;
            border-left: 5px solid #dc3545;
        }
        .stat-group {
            margin-bottom: 15px;
            font-size: 1.3rem;
            font-weight: bold;
            color: #2c3e50;
            border-bottom: 2px solid #eaeaea;
            padding-bottom: 8px;
        }
        header {
            margin-bottom: 30px;
            border-bottom: 2px solid #eaeaea;
            padding-bottom: 20px;
            text-align: center;
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 2.5rem;
        }
        .subtitle {
            color: #7f8c8d;
            font-style: italic;
            font-size: 1.2rem;
        }
        .response-times {
            grid-column: 1 / -1;
            background-color: #ebf4ff;
            border-top: 4px solid #3182ce;
        }
        .response-card {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 15px;
            box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
            border-left: 4px solid #3182ce;
        }
        
        /* Color-coded categories */
        .server-overview-card {
            border-top-color: #4299e1; /* Blue */
        }
        .memory-cpu-card {
            border-top-color: #48bb78; /* Green */
        }
        .request-stats-card {
            border-top-color: #ed8936; /* Orange */
        }
        .capacity-card {
            border-top-color: #9f7aea; /* Purple */
        }
        
        /* Making values more readable */
        .emphasized {
            color: #4299e1;
            font-weight: 700;
        }
        
        /* Animated refresh indicator */
        .refresh-indicator {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background-color: rgba(0,0,0,0.7);
            color: white;
            padding: 8px 15px;
            border-radius: 30px;
            font-size: 0.9rem;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .refresh-dot {
            height: 10px;
            width: 10px;
            background-color: #4caf50;
            border-radius: 50%;
            display: inline-block;
            animation: pulse 1s infinite;
        }
        @keyframes pulse {
            0% { opacity: 0.5; }
            50% { opacity: 1; }
            100% { opacity: 0.5; }
        }
        
        /* Responsive improvements */
        @media (max-width: 768px) {
            .stats-dashboard {
                grid-template-columns: 1fr;
            }
            .stat-value {
                font-size: 1.5rem;
            }
            h1 {
                font-size: 2rem;
            }
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
    <div id="stats-container" hx-get="/stats/data" hx-trigger="load, every 1s" hx-swap="innerHTML">
        {{template "statsData" .}}
    </div>
    
    <!-- Refresh indicator -->
    <div class="refresh-indicator">
        <span class="refresh-dot"></span>
        <span>Updating live</span>
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

        // Update server status initially and every 2 seconds
        updateServerStatus();
        setInterval(updateServerStatus, 2000);
    </script>
</body>
</html>`

	const statsDataHTML = `<div class="stats-dashboard">
    <!-- Server overview -->
    <div class="stat-card server-overview-card">
        <div class="stat-group">Server Overview</div>
        <div class="stat-name">Uptime</div>
        <div class="stat-value emphasized">{{.uptime}}</div>
    </div>
    
    <div class="stat-card memory-cpu-card">
        <div class="stat-group">Memory Usage</div>
        <div class="stat-name">Current Memory</div>
        <div class="stat-value emphasized">{{.memory_usage}}</div>
    </div>
    
    <div class="stat-card memory-cpu-card">
        <div class="stat-group">CPU Usage</div>
        <div class="stat-name">Current CPU</div>
        <div class="stat-value emphasized">{{.cpu_usage}}</div>
    </div>
    
    <!-- Request statistics -->
    <div class="stat-card request-stats-card">
        <div class="stat-group">Request Statistics</div>
        <div class="stat-name">Total Requests</div>
        <div class="stat-value emphasized">{{.requests_total}}</div>
    </div>
    
    <div class="stat-card request-stats-card">
        <div class="stat-group">Request Success</div>
        <div class="stat-name">Succeeded</div>
        <div class="stat-value emphasized">{{.requests_succeeded}}</div>
    </div>
    
    <div class="stat-card request-stats-card">
        <div class="stat-group">Request Failures</div>
        <div class="stat-name">Failed</div>
        <div class="stat-value emphasized">{{.requests_failed}}</div>
    </div>
    
    <div class="stat-card request-stats-card">
        <div class="stat-group">Request Rate</div>
        <div class="stat-name">Requests Per Second</div>
        <div class="stat-value emphasized">{{.requests_per_second}}</div>
    </div>
    
    <div class="stat-card request-stats-card">
        <div class="stat-group">Success Rate</div>
        <div class="stat-name">Request Success Rate</div>
        <div class="stat-value emphasized">{{.success_rate}}</div>
    </div>
    
    <!-- Capacity information -->
    <div class="stat-card capacity-card">
        <div class="stat-group">Server Capacity</div>
        <div class="stat-name">Current Concurrent Requests</div>
        <div class="stat-value emphasized">{{.concurrent_requests}}</div>
    </div>
    
    <div class="stat-card capacity-card">
        <div class="stat-group">Maximum Capacity</div>
        <div class="stat-name">Max Concurrent</div>
        <div class="stat-value emphasized">{{.max_concurrent}}</div>
    </div>
    
    <div class="stat-card capacity-card">
        <div class="stat-group">Server Load</div>
        <div class="stat-name">Current Load</div>
        <div class="stat-value emphasized">{{.server_load}}</div>
    </div>
    
    <!-- Response time metrics in a wider card -->
    <div class="stat-card response-times">
        <div class="stat-group">Response Time Metrics</div>
        <div class="response-card">
            <div class="stat-name">50th Percentile (P50)</div>
            <div class="stat-value emphasized">{{.p50_response_time}}</div>
        </div>
        <div class="response-card">
            <div class="stat-name">90th Percentile (P90)</div>
            <div class="stat-value emphasized">{{.p90_response_time}}</div>
        </div>
        <div class="response-card">
            <div class="stat-name">99th Percentile (P99)</div>
            <div class="stat-value emphasized">{{.p99_response_time}}</div>
        </div>
    </div>
</div>`

	// Create the template
	var err error
	StatsTemplate = template.New("stats")
	
	// Parse the main template first
	_, err = StatsTemplate.Parse(statsHTML)
	if err != nil {
		log.Fatalf("Failed to parse stats template: %v", err)
	}
	
	// Parse the data template
	_, err = StatsTemplate.New("statsData").Parse(statsDataHTML)
	if err != nil {
		log.Fatalf("Failed to parse statsData template: %v", err)
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
