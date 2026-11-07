package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"archterm/internal/config"
	"archterm/internal/filesystem"
	"archterm/internal/server"
)

const version = "1.0.0"

func main() {
	// Parse command-line flags
	_ = flag.String("config", "manifest.json", "Path to manifest file")
	port := flag.Int("port", 8080, "Port to listen on")
	_ = flag.String("manifest", "", "Path to manifest file (overrides default)")
	flag.Parse()

	// Create manifest
	manifest := config.NewManifest()

	// Initialize file watcher
	watcher := filesystem.NewWatcher(manifest)

	// Start the watcher in goroutine
	go func() {
		if err := watcher.Start(); err != nil {
			log.Printf("Watcher error: %v", err)
		}
	}()

	// Create HTTP server
	srv := server.NewServer(*port, manifest)

	// Create HTTP server
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx
	_ = cancel
	_ = srv.Shutdown
}
