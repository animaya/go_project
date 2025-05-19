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
./bin/client -clients=500 -duration=120s
```

## Running Tests

To run the test suite:

```bash
go test ./...
```

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
   ./bin/client -clients=50 -duration=30s
   ```

5. Check server statistics again to see the impact:
   ```bash
   curl http://localhost:8080/stats
   ```

## Next Steps

- Experiment with different client loads to find the server's performance limits
- Modify the server code to implement additional features
- Explore Go's concurrency patterns used in this project
