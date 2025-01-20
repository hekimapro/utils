package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// extractDatabaseName parses the given database URI and retrieves the database name.
// Returns the database name as a string or an error if the extraction fails.
func extractDatabaseName(databaseURI string) (string, error) {
	// Parse the database URI.
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		return "", fmt.Errorf("invalid database URI provided: %w", err)
	}

	// Extract the path segments from the URI.
	pathParts := strings.Split(parsedURI.Path, "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		return "", fmt.Errorf("database name not found in URI")
	}

	// Return the extracted database name.
	return pathParts[1], nil
}

// ConnectToDatabase establishes a connection to the PostgreSQL database.
// Parameters:
// - databaseURI: A string representing the URI for the database connection.
// Returns:
// - A pointer to a sql.DB instance if the connection is successful.
// - An error if the connection fails.
func ConnectToDatabase(databaseURI string) (*sql.DB, error) {
	// Extract the database name from the URI for logging purposes.
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		log.Printf("[ERROR] Failed to extract database name: %v", err)
		return nil, err
	}

	// Attempt to open a connection to the PostgreSQL database.
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		log.Printf("[ERROR] Failed to open database connection: %v", err)
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Verify the database connection by pinging it.
	if err := db.Ping(); err != nil {
		log.Printf("[ERROR] Failed to ping database: %v", err)
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	log.Printf("[INFO] Successfully connected to the database: %s", databaseName)
	return db, nil
}
