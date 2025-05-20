package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amirahmetzanov/go_project/internal/server"
)

func main() {
	// Create a server with default options
	options := server.DefaultServerOptions()
	srv := server.NewServer(options)
	
	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start the server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
	
	log.Println("Server is ready to handle requests")
	
	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")
	
	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error during server shutdown: %v", err)
	}
}
