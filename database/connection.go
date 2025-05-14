package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// extractDatabaseName parses a database URI to extract the database name
// Returns the database name or an error if the URI is invalid or lacks a name
func extractDatabaseName(databaseURI string) (string, error) {
	// Parse the database URI
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		return "", fmt.Errorf("invalid database URI provided: %w", err)
	}
	// Split the path to extract the database name
	pathParts := strings.Split(parsedURI.Path, "/")
	// Validate that a database name is present
	if len(pathParts) < 2 || pathParts[1] == "" {
		return "", fmt.Errorf("database name not found in URI")
	}
	// Return the database name
	return pathParts[1], nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database
// Configures connection pooling and verifies connectivity
// Returns the database handle or an error if the connection fails
func ConnectToDatabase(databaseURI string) (*sql.DB, error) {
	// Extract the database name from the URI
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		// Log and return error if extraction fails
		log.Printf("[ERROR] Failed to extract database name: %v", err)
		return nil, err
	}

	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		// Log and return wrapped error if connection opening fails
		log.Printf("[ERROR] Failed to open database connection: %v", err)
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Configure connection pool for high concurrency
	db.SetMaxOpenConns(450)                 // Set maximum open connections to 450
	db.SetMaxIdleConns(50)                  // Set maximum idle connections to 50
	db.SetConnMaxLifetime(2 * time.Hour)    // Set connection lifetime to 2 hours
	db.SetConnMaxIdleTime(15 * time.Minute) // Set idle connection timeout to 15 minutes

	// Verify database connectivity with a ping
	if err := db.Ping(); err != nil {
		// Log and return wrapped error if ping fails
		log.Printf("[ERROR] Failed to ping database: %v", err)
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	// Log successful connection
	log.Printf("[INFO] Successfully connected to the database: %s", databaseName)
	// Return the database handle
	return db, nil
}
