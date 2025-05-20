# How to Use the Name Generator Web Server

This guide provides a simple step-by-step tutorial on how to build, run, and test the Name Generator Web Server project.

## Building the Project

First, let's build both the server and client components:

```bash
# Navigate to the project directory
cd /Users/amirahmetzanov/go/go_project

# Build the server and client using the Makefile
make build
```

## Running the Server

To run the server:

```bash
./bin/server
```

This will start the server on port 8080 by default.

## Testing with curl

You can test the server's functionality using curl:

```bash
# Test the name generation endpoint
curl -X POST -H "Content-Type: application/json" -d '{"session_id":"test-123", "letter":"A", "num_of_entries":3}' http://localhost:8080/generate

# Check the server statistics
curl http://localhost:8080/stats
```

## Running the Client Simulator

The client simulator can generate load to test the server's performance:

```bash
# Run with default settings (100 concurrent clients for 60 seconds)
./bin/client

# Run with custom settings
./bin/client -clients=500 -duration=120s -ramp-up=30s -stats-interval=10s
```

### Client Simulator Options

- `-url`: Server URL (default: http://localhost:8080/generate)
- `-clients`: Number of concurrent clients (default: 100)
- `-duration`: Test duration (default: 60s)
- `-ramp-up`: Ramp-up duration to gradually start clients (default: 5s)
- `-stats-interval`: Interval for printing statistics (default: 5s)

## Running Tests

To run the test suite:

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./internal/generator
go test ./internal/cache
go test ./internal/ratelimit
go test ./internal/metrics
go test ./internal/workerpool
go test ./internal/server
```

## Running the Demo Script

For a quick demonstration of the server's capabilities:

```bash
./run_demo.sh
```

This script:
1. Builds both the server and client
2. Starts the server
3. Tests it with a simple curl request
4. Runs the client simulator with increasing load
5. Displays server statistics at each stage

## Example Workflow

Here's a complete example workflow:

1. Start the server in one terminal:
   ```bash
   ./bin/server
   ```

2. In another terminal, run a quick curl test:
   ```bash
   curl -X POST -H "Content-Type: application/json" -d '{"session_id":"test-123", "letter":"B", "num_of_entries":5}' http://localhost:8080/generate
   ```

3. Check server statistics:
   ```bash
   curl http://localhost:8080/stats
   ```

4. Run the client simulator with moderate load:
   ```bash
   ./bin/client -clients=50 -duration=30s -ramp-up=10s
   ```

5. Check server statistics again to see the impact:
   ```bash
   curl http://localhost:8080/stats
   ```

## Performance Testing

To test the server's performance limits:

1. Start with a moderate load:
   ```bash
   ./bin/client -clients=100 -duration=30s
   ```

2. Gradually increase the number of clients to find the throughput limit:
   ```bash
   ./bin/client -clients=200 -duration=30s
   ./bin/client -clients=500 -duration=30s
   ./bin/client -clients=1000 -duration=30s
   ```

3. Analyze the stats after each test to see how the server handles increasing load.

4. Look for these indicators in the server stats:
   - **Requests per second**: Maximum throughput
   - **Response times (p50, p90, p99)**: How latency increases under load
   - **Server load**: How close the server is to its capacity
   - **Memory usage**: How memory consumption changes under load
   - **Success rate**: Whether requests are being processed correctly

## Advanced Usage

### Worker Pool Configuration

The server uses a worker pool for name generation. You can modify the worker pool size in the server code:

```go
// In internal/server/server.go:
nameGenerator := generator.NewNameGenerator(8) // Increase from default 4
```

### Rate Limiting

The server implements sophisticated rate limiting. You can adjust the rate limits in the server options:

```go
// In cmd/server/main.go:
options := server.DefaultServerOptions()
options.RequestRateLimit = 1000 // Increase from default 500
srv := server.NewServer(options)
```

### Caching

The server uses an LRU cache for frequently requested name combinations. You can adjust the cache settings:

```go
// In cmd/server/main.go:
options := server.DefaultServerOptions()
options.CacheSize = 2000         // Increase from default 1000
options.CacheExpiration = 10 * time.Minute // Increase from default 5 minutes
srv := server.NewServer(options)
```

## Next Steps

- Experiment with different client loads to find the server's performance limits
- Modify the server code to implement additional features
- Explore Go's concurrency patterns used in this project
- Try running multiple server instances and load balancing between them
