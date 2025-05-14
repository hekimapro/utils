package server

import (
	"context"
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

// StartServer initializes and starts an HTTP or HTTPS server with graceful shutdown
// Runs the server based on the environment and handles context cancellation
// Returns an error if the server fails to start or encounters issues
func StartServer(ctx context.Context, router *chi.Mux, port, SSLKeyPath, SSLCertPath string) error {
	// Determine the environment based on SSL file presence
	env := determineEnvironment(SSLKeyPath, SSLCertPath)

	// Configure the HTTP server with timeouts, header limits, and logging
	server := &http.Server{
		Handler:        router,                       // Use the provided Chi router
		Addr:           ":" + port,                   // Set server address with port
		ReadTimeout:    30 * time.Second,             // Set read timeout
		WriteTimeout:   30 * time.Second,             // Set write timeout
		IdleTimeout:    10 * time.Second,             // Set idle timeout
		MaxHeaderBytes: 1 << 20,                      // Set maximum header size (1MB)
		ErrorLog:       log.New(os.Stderr, "[ERROR] ", log.LstdFlags), // Configure error logging
	}

	// Create a channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine to allow concurrent error handling
	go func() {
		// Log server startup details
		log.Printf("[INFO] %s server is running on port %s", env, port)

		var err error
		if env == "Development" {
			// Start an HTTP server for development
			err = server.ListenAndServe()
		} else {
			// Configure TLS for production
			tlsConfig := &tls.Config{}
			// Load SSL certificate and key
			cert, err := tls.LoadX509KeyPair(SSLCertPath, SSLKeyPath)
			if err != nil {
				serverErrors <- err
				return
			}
			tlsConfig.Certificates = []tls.Certificate{cert}

			// Create a TLS listener for the server
			listener, err := tls.Listen("tcp", server.Addr, tlsConfig)
			if err != nil {
				serverErrors <- err
				return
			}
			// Start the HTTPS server
			err = server.Serve(listener)
			if err != nil {
				serverErrors <- err
				return
			}
		}

		// Send non-closing errors to the error channel
		if err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Handle context cancellation or server errors
	select {
	case <-ctx.Done():
		// Initiate graceful shutdown on context cancellation
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Log shutdown initiation
		log.Println("[INFO] Shutting down server...")
		// Perform graceful shutdown and return any error
		return server.Shutdown(shutdownCtx)

	case err := <-serverErrors:
		// Return any error from server startup or runtime
		return err
	}
}