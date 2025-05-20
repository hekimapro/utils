package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/hekimapro/utils/log"
)

// determineEnvironment returns "Production" if both SSL cert and key files exist, otherwise "Development"
func determineEnvironment(sslKeyPath, sslCertPath string) string {
	if _, err := os.Stat(sslKeyPath); errors.Is(err, os.ErrNotExist) {
		return "Development"
	}
	if _, err := os.Stat(sslCertPath); errors.Is(err, os.ErrNotExist) {
		return "Development"
	}
	return "Production"
}

// StartServer starts an HTTP or HTTPS server with graceful shutdown support
func StartServer(ctx context.Context, router *chi.Mux, port, sslKeyPath, sslCertPath string) error {
	env := determineEnvironment(sslKeyPath, sslCertPath)

	server := &http.Server{
		Handler:        router,
		Addr:           ":" + port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Success(fmt.Sprintf("%s server is running on port %s", env, port))

		var err error
		if env == "Development" {
			err = server.ListenAndServe()
		} else {
			tlsConfig := &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				},
				NextProtos: []string{"h2", "http/1.1"}, // Enable HTTP/2 and HTTP/1.1 ALPN
			}

			// Load SSL cert & key (correct order: cert then key)
			cert, loadErr := tls.LoadX509KeyPair(sslCertPath, sslKeyPath)
			if loadErr != nil {
				serverErrors <- loadErr
				return
			}
			tlsConfig.Certificates = []tls.Certificate{cert}

			listener, listenErr := tls.Listen("tcp", server.Addr, tlsConfig)
			if listenErr != nil {
				serverErrors <- listenErr
				return
			}

			err = server.Serve(listener)
		}

		if err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)

	case err := <-serverErrors:
		return err
	}
}
