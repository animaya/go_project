#!/bin/bash
echo "Building and restarting the server..."

# Navigate to the project directory
cd /Users/amirahmetzanov/go/go_project

# Build the server
go build -o bin/server cmd/server/main.go

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Build failed! Please check errors above."
    exit 1
fi

# Kill any existing server process
pkill -f "bin/server" || true

# Start the server
./bin/server &

echo "Server restarted successfully! Access at http://localhost:8080/stats"
