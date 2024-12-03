package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"loadbalancer/internal/balancer"
	"loadbalancer/internal/config"
	"loadbalancer/internal/metrics"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize metrics
	metrics := metrics.New()

	// Create and start the load balancer
	lb, err := balancer.New(cfg, metrics)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	// Start the load balancer
	if err := lb.Start(ctx); err != nil {
		log.Fatalf("Load balancer error: %v", err)
	}
}
