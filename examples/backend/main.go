package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 9001, "Port to listen on")
	id := flag.String("id", "1", "Backend ID")
	flag.Parse()

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "healthy")
	})

	// Main handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK - Response from backend %s (ID: %s) on host %s\n", *id, *id, hostname)
		
		// Log the request for monitoring
		log.Printf("Request handled by backend %s: %s %s", *id, r.Method, r.URL.Path)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Backend %s starting on %s", *id, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
