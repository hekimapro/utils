package utils

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// extractDatabaseName parses the database URI to extract the database name.
func extractDatabaseName(databaseURI string) (string, error) {
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		return "", CreateError("invalid database URI")
	}
	pathParts := strings.Split(parsedURI.Path, "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		return "", CreateError("database name not found in URI")
	}
	return pathParts[1], nil
}

// Connect establishes a connection to the PostgreSQL database.
// It takes the database URI as a parameter and returns a sql.DB instance or an error.
func Connect(databaseURI string) (*sql.DB, error) {
	// Extract the database name from the URI.
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		log.Printf("Error extracting database name: %v", err)
		return nil, err
	}

	// Open a connection to the PostgreSQL database using the provided URI.
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		return nil, CreateError("failed to open database connection")
	}

	// Test the database connection.
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, CreateError("failed to ping database")
	}

	log.Printf("Database '%s' is successfully connected.", databaseName)
	return db, nil
}
