package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"archterm/internal/config"
	"archterm/internal/filesystem"
	"archterm/internal/network"
	"archterm/internal/server"
)

const version = "1.0.0"

func main() {
	// Parse command-line flags
	_ = flag.String("config", "manifest.json", "Path to manifest file")
	port := flag.Int("port", 8080, "Port to listen on")
	bindAddr := flag.String("bind", "", `Address to bind to (e.g., ":8080" for all interfaces, "127.0.0.1:8080" for localhost only). Default: all interfaces.`)
	showBanner := flag.Bool("banner", true, "Show Tailscale/network info banner on startup")
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

	// Print network startup banner
	if *showBanner {
		network.PrintStartupBanner(*port)
	}

	// Determine bind address
	var addr string
	if *bindAddr != "" {
		// User specified a bind address
		addr = *bindAddr
		// If there's no port, prepend one
		if !strings.Contains(addr, ":") {
			addr = addr + ":" + strconv.Itoa(*port)
		}
	} else {
		// Default: bind to all interfaces (works for Tailscale)
		addr = ":" + strconv.Itoa(*port)
	}

	log.Printf("Binding to: %s", addr)

	// Create HTTP server
	srv := server.NewServer(*port, manifest)

	// Start the server
	go func() {
		if err := srv.Start(addr); err != nil {
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
