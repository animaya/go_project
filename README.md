# Go Project: Name Generator Web Server

A high-performance, concurrent web server built in Go that generates random names based on specified criteria. This project demonstrates Go's concurrency mechanisms and scalability features.

## Overview

This project consists of two main components:
1. **Web Server**: A Go HTTP server that generates random names starting with specified letters
2. **Client Simulator**: A load testing tool that simulates thousands of concurrent clients

## Features

- High-performance name generation using Go's concurrency
- Detailed server statistics and monitoring
- Request rate limiting to prevent server overload
- Client load testing with performance metrics
- Graceful server shutdown

## Project Structure

```
go_project/
├── cmd/
│   ├── server/         # Server implementation
│   │   └── main.go
│   └── client/         # Client simulator
│       └── main.go
├── internal/
│   ├── generator/      # Name generation logic
│   │   └── generator.go
│   └── stats/          # Statistics tracking
│       └── stats.go
└── PRD.md              # Product Requirements Document
```

## Requirements

- Go 1.18 or higher

## Building the Project

To build the server:

```bash
cd /Users/amirahmetzanov/go/go_project
go build -o bin/server cmd/server/main.go
```

To build the client simulator:

```bash
cd /Users/amirahmetzanov/go/go_project
go build -o bin/client cmd/client/main.go
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
# Run directly
go run cmd/client/main.go -clients=100 -duration=60s

# Or run the built binary
./bin/client -clients=100 -duration=60s
```

### Client Simulator Options

- `-url`: Server URL (default: http://localhost:8080/generate)
- `-clients`: Number of concurrent clients (default: 100)
- `-duration`: Test duration (default: 60s)

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
### requests_processed - 2349
### memory used - 23 MB
### capacity of this server for processing requests - 3/10
```

## Performance Considerations

- The server uses a token-based rate limiter to manage concurrent requests
- Go's goroutines handle concurrent name generation
- Atomic operations ensure accurate statistics
- Memory usage is monitored in real time

## Future Improvements

- Distributed server architecture
- Persistent storage for statistics
- More sophisticated name generation algorithms
- Enhanced load balancing capabilities
