package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Define the apiConfig struct to keep track of state
type apiConfig struct {
	fileserverHits int
}

// Middleware to increment the counter for file server requests
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

// Handler for serving the metrics
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits)
}

// Handler for resetting the counter
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Hits reset to 0")
}

func main() {
	apiCfg := &apiConfig{fileserverHits: 0}
	mux := http.NewServeMux()

	// Diagnostic check to ensure directory exists
	if _, err := os.Stat("./static"); os.IsNotExist(err) {
		log.Fatalf("Static directory not found: %v", err)
	}

	// Additional Diagnostic: Print directory contents
	files, err := os.ReadDir("./static")
	if err != nil {
		log.Fatalf("Failed to read static directory: %v", err)
	}
	log.Println("Contents of ./static directory:")
	for _, file := range files {
		log.Println(" - ", file.Name())
	}

	// Create the file server handler pointing to the correct directory
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir("./static")))

	// Wrap the file server handler with the middleware
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	// Register the metrics handler
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)
	// Register the reset handler
	mux.Handle("/reset", http.HandlerFunc(apiCfg.resetHandler))

	// Start the server on port 8080
	port := 8080
	log.Printf("Starting server on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}