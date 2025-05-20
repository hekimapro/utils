package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hekimapro/utils/log"
	_ "github.com/lib/pq"
)

// extractDatabaseName parses a database URI to extract the database name
// Returns the database name or an error if the URI is invalid or lacks a name
func extractDatabaseName(databaseURI string) (string, error) {
	log.Info("Parsing database URI to extract database name")

	// Parse the database URI
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		log.Error(fmt.Sprintf("Invalid database URI: %v", err))
		return "", fmt.Errorf("invalid database URI provided: %w", err)
	}

	// Split the path to extract the database name
	pathParts := strings.Split(parsedURI.Path, "/")

	// Validate that a database name is present
	if len(pathParts) < 2 || pathParts[1] == "" {
		log.Error("Database name not found in URI")
		return "", fmt.Errorf("database name not found in URI")
	}

	dbName := pathParts[1]
	log.Success(fmt.Sprintf("Database name extracted: %s", dbName))
	return dbName, nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database
// Configures connection pooling and verifies connectivity
// Returns the database handle or an error if the connection fails
func ConnectToDatabase(databaseURI string) (*sql.DB, error) {
	log.Info("Starting database connection process")

	// Extract the database name from the URI
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to extract database name: %v", err))
		return nil, err
	}

	log.Info("Opening connection to PostgreSQL database")
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to open database connection: %v", err))
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Configure connection pool settings
	log.Info("Configuring database connection pool")
	db.SetMaxOpenConns(450)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(2 * time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)

	// Verify database connectivity
	log.Info("Pinging database to verify connectivity")
	if err := db.Ping(); err != nil {
		log.Error(fmt.Sprintf("Failed to ping database: %v", err))
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	log.Success(fmt.Sprintf("Successfully connected to the database: %s", databaseName))
	return db, nil
}
