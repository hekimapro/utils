package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

// DetermineEnvironment checks the presence of SSL certificate and key files
// to determine if the application is running in a production or development environment.
func determineEnvironment(SSLKeyPath, SSLCertPath string) string {
	if _, err := os.Stat(SSLKeyPath); os.IsNotExist(err) {
		return "Development"
	}
	if _, err := os.Stat(SSLCertPath); os.IsNotExist(err) {
		return "Development"
	}
	return "Production"
}

// StartServer initializes and starts an HTTP or HTTPS server based on the environment.
func StartServer(Router *chi.Mux, PORT, SSLKeyPath, SSLCertPath string) {
	// Determine the environment (Development or Production).
	environment := determineEnvironment(SSLKeyPath, SSLCertPath)

	// Configure the HTTP server.
	server := &http.Server{
		Handler:        Router,
		MaxHeaderBytes: 1 << 30,
		Addr:           ":" + PORT,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    10 * time.Second,
		ErrorLog:       log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}

	// Log server startup message.
	log.Printf("[INFO] %s server is running on port %s", environment, PORT)

	// Start the server based on the environment.
	if environment == "Development" {
		// Start the HTTP server for development.
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("[ERROR] HTTP server failed to start: %v", err)
		}
	} else {
		// Start the HTTPS server for production.
		tlsConfig := &tls.Config{}

		// Load SSL certificate and key.
		certificate, err := tls.LoadX509KeyPair(SSLCertPath, SSLKeyPath)
		if err != nil {
			log.Fatalf("[ERROR] Failed to load SSL certificate and key: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}

		// Create a TLS listener.
		tlsListener, err := tls.Listen("tcp", server.Addr, tlsConfig)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create TLS listener: %v", err)
		}

		// Start the HTTPS server.
		if err := server.Serve(tlsListener); err != nil {
			log.Fatalf("[ERROR] HTTPS server failed to start: %v", err)
		}
	}
}
