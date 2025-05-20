FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install make
RUN apk add --no-cache make

# Copy the go module files first and download dependencies
# This will be cached if the go.mod and go.sum files don't change
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN make build

# Use a slim image for the runtime
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binaries from the builder stage
COPY --from=builder /app/bin/server /app/bin/server

# Expose the application port
EXPOSE 8080

# Run the server
CMD ["/app/bin/server"]
