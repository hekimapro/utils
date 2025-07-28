package server

import (
	"context"    // context provides support for cancellation and timeouts.
	"crypto/tls" // tls provides support for TLS configuration and certificates.
	"errors"     // errors provides utilities for error handling.
	"fmt"        // fmt provides formatting and printing functions.
	"net/http"   // http provides HTTP server functionality.
	"os"         // os provides file system operations for checking SSL files.
	"os/signal"
	"runtime" // runtime provides access to system resources like CPU count.
	"syscall"
	"time" // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/helpers"
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// determineEnvironment returns "Production" if both SSL cert and key files exist, otherwise "Development".
// Checks the presence of SSL certificate and key files to determine the server mode.
func determineEnvironment(sslKeyPath, sslCertPath string) string {
	// Check if the SSL key file exists.
	if _, err := os.Stat(sslKeyPath); errors.Is(err, os.ErrNotExist) {
		// Log warning and return Development mode if key file is missing.
		log.Warning("SSL key file not found: running in Development mode")
		return "Development"
	}
	// Check if the SSL certificate file exists.
	if _, err := os.Stat(sslCertPath); errors.Is(err, os.ErrNotExist) {
		// Log warning and return Development mode if certificate file is missing.
		log.Warning("SSL certificate file not found: running in Development mode")
		return "Development"
	}
	// Log and return Production mode if both files are found.
	log.Info("SSL certificate and key found: running in Production mode")
	return "Production"
}

// StartServer starts an HTTP or HTTPS server with graceful shutdown support.
// Uses the provided hanlder and port, and supports TLS for Production mode.
// Returns an error if the server fails to start or encounters issues during operation.
func StartServer(handler http.Handler) error {

	// Set the number of OS threads to the number of CPU cores for optimal performance.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := helpers.GetENVValue("port")
	sslKeyPath := helpers.GetENVValue("ssl key path")
	sslCertPath := helpers.GetENVValue("ssl cert path")

	if port == "" {
		port = "8080"
		log.Warning(".env PORT is not set, defaulting to 8080")
	}

	// Determine the server environment (Production or Development).
	env := determineEnvironment(sslKeyPath, sslCertPath)

	// Log the server startup details.
	log.Info(fmt.Sprintf("Starting %s server on port %s", env, port))

	// Configure the HTTP server with timeouts and header limits.
	server := &http.Server{
		Handler:        handler,
		Addr:           ":" + port,       // Bind to the specified port.
		ReadTimeout:    30 * time.Second, // Set read timeout to 30 seconds.
		WriteTimeout:   30 * time.Second, // Set write timeout to 30 seconds.
		IdleTimeout:    10 * time.Second, // Set idle timeout to 10 seconds.
		MaxHeaderBytes: 1 << 20,          // Limit header size to 1MB.
	}

	// Create a channel to receive server errors.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine to handle HTTP or HTTPS based on environment.
	go func() {
		var err error
		if env == "Development" {
			// Start an HTTP server in Development mode.
			log.Info("Launching HTTP server (Development)")
			err = server.ListenAndServe()
		} else {
			// Start an HTTPS server in Production mode with TLS.
			log.Info("Launching HTTPS server (Production) with TLS")

			// Configure TLS with secure settings.
			tlsConfig := &tls.Config{
				MinVersion:               tls.VersionTLS12, // Enforce TLS 1.2 or higher.
				PreferServerCipherSuites: true,             // Prefer server-selected cipher suites.
				CurvePreferences: []tls.CurveID{ // Specify preferred elliptic curves.
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				},
				NextProtos: []string{"h2", "http/1.1"}, // Support HTTP/2 and HTTP/1.1.
			}

			// Load the SSL certificate and key pair.
			cert, loadErr := tls.LoadX509KeyPair(sslCertPath, sslKeyPath)
			if loadErr != nil {
				// Log and send error if certificate loading fails.
				log.Error("Failed to load SSL cert and key: " + loadErr.Error())
				serverErrors <- loadErr
				return
			}
			tlsConfig.Certificates = []tls.Certificate{cert}

			// Create a TLS listener for the server.
			listener, listenErr := tls.Listen("tcp", server.Addr, tlsConfig)
			if listenErr != nil {
				// Log and send error if TLS listener creation fails.
				log.Error("Failed to start TLS listener: " + listenErr.Error())
				serverErrors <- listenErr
				return
			}

			// Start the HTTPS server.
			err = server.Serve(listener)
		}

		// Send any server errors (except graceful shutdown) to the error channel.
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error: " + err.Error())
			serverErrors <- err
		}
	}()

	// Wait for either a context cancellation or a server error.
	select {
	case <-ctx.Done():
		// Handle graceful shutdown on context cancellation.
		log.Info("Received shutdown signal, shutting down server gracefully...")
		// Create a timeout context for shutdown (10 seconds).
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Attempt to shut down the server gracefully.
		if err := server.Shutdown(shutdownCtx); err != nil {
			// Log and return an error if shutdown fails.
			log.Error("Error during server shutdown: " + err.Error())
			return err
		}
		// Log successful shutdown.
		log.Success("Server shutdown completed successfully")
		return nil

	case err := <-serverErrors:
		// Return any server error received from the goroutine.
		return err
	}
}

func ChainMiddlewares(finalHandler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		currentMiddleware := middlewares[i]
		finalHandler = currentMiddleware(finalHandler)
	}
	return finalHandler
}
