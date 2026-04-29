// BirdNET-Go: A BirdNET-based bird call identification system
// This is the main entry point for the application.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/logger"
	"github.com/tphakala/birdnet-go/cmd"
)

// version is set at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Initialize the root context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	// Also handle SIGHUP so the process can be cleanly stopped by init systems
	// Note: on my Raspberry Pi, SIGTERM is the primary signal sent by systemd
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	// Load application configuration
	cfg, err := conf.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	appLogger, err := logger.New(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Sync() //nolint:errcheck

	appLogger.Info("Starting BirdNET-Go",
		"version", version,
		"commit", commit,
		"date", date,
	)

	// Execute the root CLI command
	if err := cmd.Execute(ctx, cfg, appLogger, version); err != nil {
		appLogger.Error("Application exited with error", "error", err)
		os.Exit(1)
	}
}
