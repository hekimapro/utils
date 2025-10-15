package server

import (
	"context"    // context provides support for cancellation and timeouts.
	"crypto/tls" // tls provides support for TLS configuration and certificates.
	"errors"     // errors provides utilities for error handling.
	"fmt"        // fmt provides formatting and printing functions.
	"net/http"   // http provides HTTP server functionality.
	"os"         // os provides file system operations for checking SSL files.
	"os/signal"  // signal provides system signal handling.
	"runtime"    // runtime provides access to system resources like CPU count.
	"strconv"    // strconv provides string conversion utilities.
	"syscall"    // syscall provides system call constants.
	"time"       // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/helpers" // helpers provides utility functions for environment variables.
	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
)

// ServerConfig holds configuration parameters for the HTTP server.
// This struct centralizes all server settings for better maintainability.
type ServerConfig struct {
	Port            string        // Port specifies the TCP port for the server to listen on
	SSLKeyPath      string        // SSLKeyPath specifies the file path to the SSL private key
	SSLCertPath     string        // SSLCertPath specifies the file path to the SSL certificate
	ReadTimeout     time.Duration // ReadTimeout is the maximum duration for reading the entire request
	WriteTimeout    time.Duration // WriteTimeout is the maximum duration for writing the response
	IdleTimeout     time.Duration // IdleTimeout is the maximum duration for idle connections
	ShutdownTimeout time.Duration // ShutdownTimeout is the duration for graceful shutdown
	MaxHeaderBytes  int           // MaxHeaderBytes limits the maximum size of request headers
	MaxConnections  int           // MaxConnections limits concurrent connections (0 = no limit)
}

// LoadConfig loads server configuration from environment variables with defaults.
// Returns a ServerConfig struct with validated and default values.
func LoadConfig() ServerConfig {
	port := helpers.GetENVValue("port")
	if port == "" {
		port = "8080"
		log.Warning(".env PORT is not set, defaulting to 8080")
	}

	return ServerConfig{
		Port:            port,
		SSLKeyPath:      helpers.GetENVValue("ssl key path"),
		SSLCertPath:     helpers.GetENVValue("ssl cert path"),
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     10 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1MB
		MaxConnections:  0,       // No limit by default
	}
}

// validatePort validates that the port is a valid TCP port number.
// Returns an error if the port is invalid or out of range.
func validatePort(port string) error {
	if port == "" {
		return errors.New("port cannot be empty")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number %d out of range (1-65535)", portNum)
	}

	return nil
}

// validateSSLPermissions validates SSL file permissions and accessibility.
// In production mode, ensures private key has proper security permissions.
// Returns an error if files are inaccessible or have insecure permissions.
func validateSSLPermissions(certPath, keyPath string) error {
	// Check if certificate file exists and is readable
	certInfo, err := os.Stat(certPath)
	if err != nil {
		return fmt.Errorf("SSL certificate file inaccessible: %w", err)
	}

	// Check if key file exists and is readable
	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		return fmt.Errorf("SSL key file inaccessible: %w", err)
	}

	// Check that key file is not world-readable (security best practice)
	if keyInfo.Mode().Perm()&0004 != 0 {
		return errors.New("SSL key file has world-readable permissions - this is insecure")
	}

	// Verify that certificate and key files are regular files
	if certInfo.Mode().IsDir() {
		return errors.New("SSL certificate path is a directory, not a file")
	}
	if keyInfo.Mode().IsDir() {
		return errors.New("SSL key path is a directory, not a file")
	}

	// Test loading the certificate pair to verify they match
	_, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("SSL certificate and key mismatch: %w", err)
	}

	return nil
}

// determineEnvironment determines the server environment based on SSL file availability.
// Returns "Production" if both SSL files exist and are valid, otherwise "Development".
// Performs comprehensive validation of SSL files in production mode.
func determineEnvironment(sslKeyPath, sslCertPath string) string {
	// Check if SSL key file exists
	if _, err := os.Stat(sslKeyPath); errors.Is(err, os.ErrNotExist) {
		log.Warning("SSL key file not found: running in Development mode")
		return "Development"
	}

	// Check if SSL certificate file exists
	if _, err := os.Stat(sslCertPath); errors.Is(err, os.ErrNotExist) {
		log.Warning("SSL certificate file not found: running in Development mode")
		return "Development"
	}

	// Validate SSL file permissions and accessibility for production mode
	if err := validateSSLPermissions(sslCertPath, sslKeyPath); err != nil {
		log.Warning("SSL file validation failed: " + err.Error() + " - running in Development mode")
		return "Development"
	}

	log.Info("SSL certificate and key validated: running in Production mode")
	return "Production"
}

// createTLSConfig creates a secure TLS configuration with modern security settings.
// Returns a *tls.Config configured with TLS 1.2+, secure ciphers, and preferred curves.
// This configuration follows security best practices for production servers.
func createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 or higher for security
		CipherSuites: []uint16{
			// Prefer ECDHE ciphers for forward secrecy
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true, // Prefer server's cipher suite order
		CurvePreferences: []tls.CurveID{
			tls.X25519, // Modern, secure curve
			tls.CurveP256,
			tls.CurveP384,
		},
		NextProtos: []string{"h2", "http/1.1"}, // Support HTTP/2 and HTTP/1.1
	}
}

// healthCheckHandler creates a basic health check endpoint handler.
// Returns an http.Handler that responds with a JSON health status.
// This provides a simple way to monitor server availability at /health.
func healthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only respond to GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON response
		response := `{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
		w.Write([]byte(response))
	})
}

// connectionLimiter creates a middleware that limits concurrent connections.
// Takes maxConnections as the maximum number of simultaneous connections.
// If maxConnections is 0, no limiting is applied.
// Returns a middleware function that enforces the connection limit.
func connectionLimiter(maxConnections int) func(http.Handler) http.Handler {
	// If maxConnections is 0, return a no-op middleware
	if maxConnections <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	semaphore := make(chan struct{}, maxConnections)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to acquire a semaphore slot
			select {
			case semaphore <- struct{}{}:
				// Release the slot when request completes
				defer func() { <-semaphore }()
				next.ServeHTTP(w, r)
			default:
				// Return 429 Too Many Requests if limit exceeded
				log.Warning("Connection limit exceeded - rejecting request")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}
		})
	}
}

// wrapHandlerWithHealthAndLimits wraps the provided handler with health endpoint and connection limiting.
// This internal function creates a new mux that includes the /health endpoint and applies connection limits.
func wrapHandlerWithHealthAndLimits(handler http.Handler, maxConnections int) http.Handler {
	// Create a new multiplexer
	mux := http.NewServeMux()

	// Register health check handler at /health
	mux.Handle("/health", healthCheckHandler())

	// Register main application handler for all other routes
	mux.Handle("/", handler)

	// Apply connection limiting if specified
	wrappedHandler := connectionLimiter(maxConnections)(mux)

	return wrappedHandler
}

// ChainMiddlewares chains multiple HTTP middlewares in the correct order.
// Takes a final handler and variadic middlewares, applying them from last to first.
// Returns the final wrapped http.Handler with all middlewares applied.
// This function maintains backward compatibility with existing projects.
func ChainMiddlewares(finalHandler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		currentMiddleware := middlewares[i]
		finalHandler = currentMiddleware(finalHandler)
	}
	return finalHandler
}

// StartServer starts an HTTP or HTTPS server with graceful shutdown support.
// Uses the provided handler and configuration, supporting TLS for Production mode.
// This function maintains backward compatibility while adding health endpoint at /health
// and optional connection limiting.
//
// Parameters:
//   - handler: The HTTP handler to serve requests (router, mux, etc.)
//
// Returns:
//   - error: Any error encountered during server startup, operation, or shutdown
//
// Example:
//
//	err := server.StartServer(router)
//	if err != nil {
//	    log.Fatal("Server failed: " + err.Error())
//	}
func StartServer(handler http.Handler) error {
	// Set the number of OS threads to the number of CPU cores for optimal performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load configuration from environment variables
	config := LoadConfig()

	// Validate port configuration
	if err := validatePort(config.Port); err != nil {
		return fmt.Errorf("port validation failed: %w", err)
	}

	// Set up context for graceful shutdown on OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Determine server environment (Production or Development)
	env := determineEnvironment(config.SSLKeyPath, config.SSLCertPath)

	// Log server startup details with configuration
	log.Info(fmt.Sprintf("Starting %s server on port %s", env, config.Port))
	log.Info(fmt.Sprintf("Timeouts - Read: %v, Write: %v, Idle: %v, Shutdown: %v",
		config.ReadTimeout, config.WriteTimeout, config.IdleTimeout, config.ShutdownTimeout))

	// Wrap the handler with health endpoint and connection limiting
	wrappedHandler := wrapHandlerWithHealthAndLimits(handler, config.MaxConnections)

	// Log connection limiting status
	if config.MaxConnections > 0 {
		log.Info(fmt.Sprintf("Connection limiting enabled: %d max concurrent connections", config.MaxConnections))
	} else {
		log.Info("Connection limiting disabled (no limit)")
	}

	// Log health endpoint availability
	log.Info("Health endpoint available at: /health")

	// Configure the HTTP server with timeouts and limits
	server := &http.Server{
		Handler:        wrappedHandler,
		Addr:           ":" + config.Port,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	// Create a channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine to handle HTTP or HTTPS based on environment
	go func() {
		var err error
		if env == "Development" {
			// Start an HTTP server in Development mode
			log.Info("Launching HTTP server (Development)")
			err = server.ListenAndServe()
		} else {
			// Start an HTTPS server in Production mode with TLS
			log.Info("Launching HTTPS server (Production) with TLS")

			// Create secure TLS configuration
			tlsConfig := createTLSConfig()

			// Load the SSL certificate and key pair
			cert, loadErr := tls.LoadX509KeyPair(config.SSLCertPath, config.SSLKeyPath)
			if loadErr != nil {
				log.Error("Failed to load SSL cert and key: " + loadErr.Error())
				serverErrors <- loadErr
				return
			}
			tlsConfig.Certificates = []tls.Certificate{cert}

			// Create a TLS listener for the server
			listener, listenErr := tls.Listen("tcp", server.Addr, tlsConfig)
			if listenErr != nil {
				log.Error("Failed to start TLS listener: " + listenErr.Error())
				serverErrors <- listenErr
				return
			}

			// Start the HTTPS server with TLS listener
			err = server.Serve(listener)
		}

		// Send any server errors (except graceful shutdown) to the error channel
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error: " + err.Error())
			serverErrors <- err
		}
	}()

	// Wait for either a context cancellation (shutdown signal) or a server error
	select {
	case <-ctx.Done():
		// Handle graceful shutdown on context cancellation
		log.Info("Received shutdown signal, shutting down server gracefully...")

		// Create a timeout context for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		// Attempt to shut down the server gracefully
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("Error during server shutdown: " + err.Error())
			return err
		}

		// Log successful shutdown
		log.Success("Server shutdown completed successfully")
		return nil

	case err := <-serverErrors:
		// Return any server error received from the goroutine
		return err
	}
}
