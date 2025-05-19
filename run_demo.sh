#!/bin/bash

# Run a demo of the name generator server and client
# This script builds and runs both the server and client components

# Function to clean up on exit
cleanup() {
    echo "Stopping server..."
    if [ -f "server.pid" ]; then
        kill $(cat server.pid) 2>/dev/null
        rm server.pid
    fi
    exit 0
}

# Set up cleanup trap
trap cleanup INT TERM EXIT

# Ensure the project is built
echo "Building project..."
make build

# Start the server in the background
echo "Starting server..."
./bin/server & echo $! > server.pid
SERVER_PID=$(cat server.pid)
echo "Server started with PID: $SERVER_PID"

# Wait for server to initialize
sleep 2

# Test the server with curl
echo "Testing server with curl..."
echo "Sending request to generate names starting with letter 'C'..."
curl -s -X POST -H "Content-Type: application/json" \
    -d '{"session_id":"demo-123", "letter":"C", "num_of_entries":5}' \
    http://localhost:8080/generate | jq .

# Check server stats
echo "Checking server statistics..."
curl -s http://localhost:8080/stats

# Run the client with a small load
echo ""
echo "Running client simulator with 10 concurrent clients for 10 seconds..."
./bin/client -clients=10 -duration=10s

# Check server stats after load test
echo "Checking server statistics after load test..."
curl -s http://localhost:8080/stats

# Give the user a chance to interact with the server
echo ""
echo "Server is running on http://localhost:8080"
echo "You can press Ctrl+C to stop the demo at any time"
echo ""
echo "Example curl commands:"
echo "curl -X POST -H \"Content-Type: application/json\" -d '{\"session_id\":\"test-456\", \"letter\":\"D\", \"num_of_entries\":3}' http://localhost:8080/generate"
echo "curl http://localhost:8080/stats"
echo ""

# Wait for the user to press Ctrl+C
wait $SERVER_PID
