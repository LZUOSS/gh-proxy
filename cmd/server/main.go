package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LZUOSS/gh-proxy/internal/config"
	"github.com/LZUOSS/gh-proxy/internal/server"
	"github.com/LZUOSS/gh-proxy/internal/ssh"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", getEnvOrDefault("CONFIG_PATH", "./configs/config.yaml"), "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully from: %s", *configPath)

	// Create HTTP server
	httpServer, err := server.NewHTTPServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// Create SSH server with a simple config
	// Using default SSH settings since SSH config is not in main config yet
	sshServer, err := ssh.NewServer(&ssh.Config{
		Address:        ":2222",
		EnablePassword: true,
		EnablePubKey:   true,
	})
	if err != nil {
		log.Fatalf("Failed to create SSH server: %v", err)
	}

	// Use WaitGroup to track server goroutines
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start HTTP server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting HTTP server...")
		if err := httpServer.Start(); err != nil {
			errChan <- err
		}
	}()

	// Start SSH server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting SSH server...")
		if err := sshServer.Start(); err != nil {
			errChan <- err
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, initiating graceful shutdown...", sig)
	case err := <-errChan:
		log.Printf("Server error: %v, initiating shutdown...", err)
	}

	// Create shutdown context with timeout
	shutdownTimeout := 30 * time.Second
	if cfg.Server.ShutdownTimeout > 0 {
		shutdownTimeout = cfg.Server.ShutdownTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown servers concurrently
	shutdownWg := sync.WaitGroup{}

	// Shutdown HTTP server
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	// Shutdown SSH server
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		if err := sshServer.Stop(); err != nil {
			log.Printf("SSH server shutdown error: %v", err)
		}
	}()

	// Wait for all shutdowns to complete or timeout
	done := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All servers stopped gracefully")
	case <-ctx.Done():
		log.Printf("Shutdown timeout exceeded (%v), forcing exit", shutdownTimeout)
	}

	log.Println("Application exited")
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
