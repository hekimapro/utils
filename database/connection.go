package database

import (
	"database/sql" // sql provides database connectivity and query execution.
	"fmt"          // fmt provides formatting and printing functions.
	"net/url"      // url provides utilities for parsing database URIs.
	"strings"      // strings provides utilities for string manipulation.
	"time"         // time provides functionality for handling connection timeouts.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
	_ "github.com/lib/pq"            // pq registers the PostgreSQL driver.
)

// extractDatabaseName parses a database URI to extract the database name.
// Returns the database name or an error if the URI is invalid or lacks a name.
func extractDatabaseName(databaseURI string) (string, error) {
	// Log the start of the URI parsing process.
	log.Info("🔍 Parsing database URI to extract database name")

	// Parse the provided database URI.
	parsedURI, err := url.Parse(databaseURI)
	if err != nil {
		// Log and return an error if the URI is invalid.
		log.Error(fmt.Sprintf("❌ Invalid database URI: %v", err))
		return "", fmt.Errorf("invalid database URI provided: %w", err)
	}

	// Split the URI path to extract the database name.
	pathParts := strings.Split(parsedURI.Path, "/")
	if len(pathParts) < 2 || pathParts[1] == "" {
		// Log and return an error if no database name is found.
		log.Error("❌ Database name not found in URI path")
		return "", fmt.Errorf("database name not found in URI")
	}

	// Extract the database name from the path.
	dbName := pathParts[1]
	// Log successful extraction of the database name.
	log.Success(fmt.Sprintf("📦 Database name extracted: %s", dbName))
	return dbName, nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database.
// Configures connection pooling and verifies connectivity.
// Returns the database handle or an error if the connection fails.
func ConnectToDatabase(databaseURI string) (*sql.DB, error) {
	// Log the start of the database connection process.
	log.Info("🔌 Starting database connection process")

	// Extract the database name from the URI.
	databaseName, err := extractDatabaseName(databaseURI)
	if err != nil {
		// Log and return an error if extraction fails.
		log.Error(fmt.Sprintf("❌ Failed to extract database name: %v", err))
		return nil, err
	}

	// Open a connection to the PostgreSQL database using the provided URI.
	log.Info("📡 Opening connection to PostgreSQL database")
	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		// Log and return an error if the connection cannot be opened.
		log.Error(fmt.Sprintf("❌ Failed to open database connection: %v", err))
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Configure connection pool settings for performance and resource management.
	log.Info("⚙️ Configuring database connection pool")
	db.SetMaxIdleConns(50)                  // Set maximum idle connections to 50.
	db.SetMaxOpenConns(500)                 // Set maximum open connections to 500.
	db.SetConnMaxLifetime(2 * time.Hour)    // Set maximum connection lifetime to 2 hours.
	db.SetConnMaxIdleTime(15 * time.Minute) // Set maximum idle time to 15 minutes.

	// Verify database connectivity with a ping.
	log.Info("🔎 Verifying database connectivity with ping")
	if err := db.Ping(); err != nil {
		// Log and return an error if the ping fails.
		log.Error(fmt.Sprintf("❌ Failed to ping database: %v", err))
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	// Log successful connection to the database.
	log.Success(fmt.Sprintf("✅ Successfully connected to database: %s", databaseName))
	return db, nil
}
