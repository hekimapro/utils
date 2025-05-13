package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

// determineEnvironment checks for SSL certificate and key files
// Returns "Production" if both files exist, otherwise "Development"
func determineEnvironment(SSLKeyPath, SSLCertPath string) string {
	// Check if the SSL key file exists
	if _, err := os.Stat(SSLKeyPath); os.IsNotExist(err) {
		return "Development"
	}
	// Check if the SSL certificate file exists
	if _, err := os.Stat(SSLCertPath); os.IsNotExist(err) {
		return "Development"
	}
	// Return Production if both files are present
	return "Production"
}

// StartServer initializes and starts an HTTP or HTTPS server
// Configures the server based on the environment and starts it
func StartServer(Router *chi.Mux, PORT, SSLKeyPath, SSLCertPath string) {
	// Determine the environment based on SSL file presence
	environment := determineEnvironment(SSLKeyPath, SSLCertPath)

	// Configure the HTTP server with timeouts and logging
	server := &http.Server{
		Handler:        Router,                       // Use the provided Chi router
		MaxHeaderBytes: 1 << 30,                      // Set maximum header size
		Addr:           ":" + PORT,                   // Set server address with port
		ReadTimeout:    30 * time.Second,             // Set read timeout
		WriteTimeout:   30 * time.Second,             // Set write timeout
		IdleTimeout:    10 * time.Second,             // Set idle timeout
		ErrorLog:       log.New(os.Stderr, "[ERROR] ", log.LstdFlags), // Configure error logging
	}

	// Log server startup details
	log.Printf("[INFO] %s server is running on port %s", environment, PORT)

	// Start the server based on the environment
	if environment == "Development" {
		// Start an HTTP server for development
		if err := server.ListenAndServe(); err != nil {
			// Log fatal error and exit if the server fails to start
			log.Fatalf("[ERROR] HTTP server failed to start: %v", err)
		}
	} else {
		// Configure TLS for production
		tlsConfig := &tls.Config{}

		// Load SSL certificate and key
		certificate, err := tls.LoadX509KeyPair(SSLCertPath, SSLKeyPath)
		if err != nil {
			// Log fatal error and exit if SSL loading fails
			log.Fatalf("[ERROR] Failed to load SSL certificate and key: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}

		// Create a TLS listener for the server
		tlsListener, err := tls.Listen("tcp", server.Addr, tlsConfig)
		if err != nil {
			// Log fatal error and exit if TLS listener creation fails
			log.Fatalf("[ERROR] Failed to create TLS listener: %v", err)
		}

		// Start the HTTPS server
		if err := server.Serve(tlsListener); err != nil {
			// Log fatal error and exit if the server fails to start
			log.Fatalf("[ERROR] HTTPS server failed to start: %v", err)
		}
	}
}