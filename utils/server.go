package utils

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
)

func getEnvironment(SSLKeyPath, SSLCertPath string) string {
	if _, err := os.Stat(SSLKeyPath); os.IsNotExist(err) {
		return "Development"
	}
	if _, err := os.Stat(SSLCertPath); os.IsNotExist(err) {
		return "Development"
	}
	return "Production"
}

// Start initializes and starts the HTTP server.
func StartServer(Router *chi.Mux, PORT, SSLKeyPath, SSLCertPath string) {

	// Get environment variables
	Environment := getEnvironment(SSLKeyPath, SSLCertPath)

	// Create HTTP server configuration
	Server := &http.Server{
		Handler:        Router,
		MaxHeaderBytes: 1 << 30,
		Addr:           ":" + PORT,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		ErrorLog:       log.New(os.Stderr, "server error: ", log.LstdFlags),
	}

	// Log server startup message
	log.Printf("%v server is running on PORT %v", Environment, PORT)

	// Start HTTP server
	if Environment == "Development" {
		// Start HTTP server without TLS in development environment
		Error := Server.ListenAndServe()
		if Error != nil {
			log.Fatal(Error.Error())
		}
	} else {
		// Start HTTPS server with TLS in production environment

		// Load SSL certificate and key
		sslKeyPath := SSLKeyPath
		sslCertPath := SSLCertPath
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		var err error
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(sslCertPath, sslKeyPath)
		if err != nil {
			log.Fatalf("Failed to load SSL certificate and key: %v", err)
		}

		// Create TLS listener
		tlsListener, err := tls.Listen("tcp", Server.Addr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to create TLS listener: %v", err)
		}

		// Start HTTPS server
		Error := Server.Serve(tlsListener)
		if Error != nil {
			log.Fatal(Error.Error())
		}
	}
}
