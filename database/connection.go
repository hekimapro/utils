package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// extractDatabaseName parses the given database URI and retrieves the database name
// Extracts the database name from the URI path for logging or reference
// Returns the database name or an error if parsing fails
func extractDatabaseName(databaseURI string) (string, error) {
	// Parse the database URI to extract its components
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		return "", fmt.Errorf("invalid database URI provided: %w", err)
	}

	// Split the URI path into segments to locate the database name
	pathParts := strings.Split(parsedURI.Path, "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		return "", fmt.Errorf("database name not found in URI")
	}

	// Return the database name from the path
	return pathParts[1], nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database
// Uses the provided database URI to open and verify a connection
// Returns a sql.DB instance for database operations or an error if connection fails
func ConnectToDatabase(databaseURI string) (*sql.DB, error) {
	// Extract the database name for logging purposes
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		// Log the error for debugging
		log.Printf("[ERROR] Failed to extract database name: %v", err)
		return nil, err
	}

	// Open a connection to the PostgreSQL database using the provided URI
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		// Log the error and return a wrapped error for context
		log.Printf("[ERROR] Failed to open database connection: %v", err)
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Verify the connection by pinging the database
	if err := db.Ping(); err != nil {
		// Log the error and return a wrapped error for context
		log.Printf("[ERROR] Failed to ping database: %v", err)
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	// Log successful connection with the database name
	log.Printf("[INFO] Successfully connected to the database: %s", databaseName)
	return db, nil
}