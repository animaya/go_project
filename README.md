# Go Project: Name Generator Web Server

A high-performance, concurrent web server built in Go that generates random names based on specified criteria. This project demonstrates Go's concurrency mechanisms and scalability features.

## Overview

This project consists of two main components:
1. **Web Server**: A Go HTTP server that generates random names starting with specified letters
2. **Client Simulator**: A load testing tool that simulates thousands of concurrent clients

## Features

- **High-performance name generation** using advanced concurrency patterns
- **Worker pool** for efficient parallel processing
- **Sophisticated rate limiting** with token bucket and sliding window algorithms
- **Distributed caching** with LRU eviction policy
- **Advanced metrics collection** with detailed performance statistics
- **Graceful server shutdown** for proper resource cleanup
- **Client load testing** with comprehensive performance metrics
- **Request pipelining** for improved throughput

## Project Structure

```
go_project/
├── cmd/
│   ├── server/         # Server implementation
│   │   └── main.go
│   └── client/         # Client simulator
│       └── main.go
├── internal/
│   ├── cache/          # Caching system
│   │   ├── cache.go
│   │   └── cache_test.go
│   ├── generator/      # Name generation logic
│   │   ├── generator.go
│   │   └── generator_test.go
│   ├── metrics/        # Performance metrics
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   ├── ratelimit/      # Rate limiting
│   │   ├── ratelimit.go
│   │   └── ratelimit_test.go
│   ├── server/         # Server implementation
│   │   ├── server.go
│   │   └── server_test.go
│   └── workerpool/     # Worker pool for parallel processing
│       ├── workerpool.go
│       └── workerpool_test.go
├── Makefile            # Build automation
├── PRD.md              # Product Requirements Document
├── README.md           # Project documentation
├── USAGE.md            # Usage instructions
└── run_demo.sh         # Demo script
```

## Requirements

- Go 1.18 or higher

## Building the Project

To build the server and client components:

```bash
cd /Users/amirahmetzanov/go/go_project
make build
```

## Running the Project

### Starting the Server

```bash
# Run directly
go run cmd/server/main.go

# Or run the built binary
./bin/server
```

The server will start on port 8080 by default.

### Running the Client Simulator

```bash
# Run with default settings (100 concurrent clients for 60 seconds)
./bin/client

# Run with custom settings
./bin/client -clients=500 -duration=2m -ramp-up=30s
```

### Client Simulator Options

- `-url`: Server URL (default: http://localhost:8080/generate)
- `-clients`: Number of concurrent clients (default: 100)
- `-duration`: Test duration (default: 60s)
- `-ramp-up`: Ramp-up duration to gradually start clients (default: 5s)
- `-stats-interval`: Interval for printing statistics (default: 5s)

## API Endpoints

### Generate Names

**Endpoint**: `POST /generate`

**Request Example:**
```json
{
  "session_id": "123-456",
  "letter": "A",
  "num_of_entries": 5
}
```

**Response Example:**
```json
{
  "session_id": "123-456",
  "names": ["Anna", "Alex", "Andrew", "Aaron", "Alice"],
  "num_of_entries": 5
}
```

### Server Statistics

**Endpoint**: `GET /stats`

**Response Example:**
```
## Web server statistics
### uptime - 1h23m45s
### requests_total - 2349
### requests_succeeded - 2349
### requests_failed - 0
### requests_per_second - 23.45
### success_rate - 100.00%
### concurrent_requests - 87
### max_concurrent - 1000
### server_load - 0.87/10
### memory_usage - 24.32 MB
### cpu_usage - 34.21%
### p50_response_time - 42ms
### p90_response_time - 78ms
### p99_response_time - 156ms
### avg_response_time - 55ms
```

## Performance Considerations

- **Worker Pool**: Efficiently processes requests in parallel using a fixed number of workers
- **Distributed Cache**: Reduces load by caching frequently requested name combinations
- **Token Bucket Rate Limiter**: Manages request rate with burst capability
- **Sliding Window Rate Limiter**: Provides additional protection against traffic spikes
- **Metrics Collection**: Tracks detailed performance statistics for monitoring
- **Graceful Shutdown**: Ensures proper cleanup of resources when the server stops

## Demo Script

For a quick demonstration of the server's capabilities, run:

```bash
./run_demo.sh
```

This script:
1. Builds both the server and client
2. Starts the server
3. Tests it with a simple curl request
4. Runs the client simulator with increasing load
5. Displays server statistics at each stage

## Future Improvements

- Distributed server architecture with load balancing
- More sophisticated name generation algorithms
- Enhanced rate limiting for distributed environments
- Dynamic worker pool sizing based on load
- Real-time monitoring dashboard
