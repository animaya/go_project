
# Makefile for the Go project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Binary names
SERVER_BIN=server
CLIENT_BIN=client

# Build directory
BIN_DIR=bin

# Source files
SERVER_SRC=cmd/server/main.go
CLIENT_SRC=cmd/client/main.go

# Default make target
all: clean build

# Create bin directory
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build the project
build: $(BIN_DIR) build-server build-client

# Build the server
build-server:
	$(GOBUILD) -o $(BIN_DIR)/$(SERVER_BIN) $(SERVER_SRC)
	@echo "Server binary built successfully!"

# Build the client
build-client:
	$(GOBUILD) -o $(BIN_DIR)/$(CLIENT_BIN) $(CLIENT_SRC)
	@echo "Client binary built successfully!"

# Clean the project
clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	@echo "Project cleaned!"

# Run the server
run-server:
	$(GORUN) $(SERVER_SRC)

# Run the client (can specify arguments with args="...")
run-client:
	$(GORUN) $(CLIENT_SRC) $(args)

# Test the project
test:
	$(GOTEST) ./...

# Run both server and client (in background and foreground respectively)
run: build
	@echo "Starting server in background..."
	@./$(BIN_DIR)/$(SERVER_BIN) & echo $$! > server.pid
	@echo "Server started with PID $$(cat server.pid)"
	@sleep 2
	@echo "Starting client..."
	./$(BIN_DIR)/$(CLIENT_BIN) -clients=100 -duration=30s
	@echo "Stopping server..."
	@kill $$(cat server.pid)
	@rm server.pid
	@echo "Server stopped"

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build the project"
	@echo "  build        - Build the project (server and client)"
	@echo "  build-server - Build only the server"
	@echo "  build-client - Build only the client"
	@echo "  clean        - Clean the project"
	@echo "  run-server   - Run the server"
	@echo "  run-client   - Run the client (can specify arguments with args=\"...\")"
	@echo "  run          - Run both server and client"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help message"

.PHONY: all build build-server build-client clean run-server run-client test run help
