package database

import (
	"database/sql" // sql provides database connectivity and query execution.
	"fmt"          // fmt provides formatting and printing functions.
	"net/url"      // url provides utilities for parsing database URIs.
	"strings"

	// strings provides utilities for string manipulation.
	"time" // time provides functionality for handling connection timeouts.

	"github.com/hekimapro/utils/env"
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
	"github.com/hekimapro/utils/models"
	_ "github.com/lib/pq" // pq registers the PostgreSQL driver.
)

func getURI(databaseOptions models.DatabaseOptions) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(databaseOptions.Username),
		url.QueryEscape(databaseOptions.Password),
		databaseOptions.Host,
		databaseOptions.Port,
		databaseOptions.DatabaseName,
		url.QueryEscape(databaseOptions.SSLMode),
	)
}

// validateDatabaseOptions checks for required fields in DatabaseOptions and returns an error if any are missing.
func validateDatabaseOptions(opts models.DatabaseOptions) error {
	var missing []string

	if strings.TrimSpace(opts.Username) == "" {
		missing = append(missing, "Username")
	}
	if strings.TrimSpace(opts.Password) == "" {
		missing = append(missing, "Password")
	}
	if strings.TrimSpace(opts.Host) == "" {
		missing = append(missing, "Host")
	}
	if strings.TrimSpace(opts.Port) == "" {
		missing = append(missing, "Port")
	}
	if strings.TrimSpace(opts.DatabaseName) == "" {
		missing = append(missing, "DatabaseName")
	}
	if strings.TrimSpace(opts.SSLMode) == "" {
		missing = append(missing, "SSLMode")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required database option(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database.
// Configures connection pooling and verifies connectivity.
// Returns the database handle or an error if the connection fails.
func ConnectToDatabase() (*sql.DB, error) {

	databaseOptions := models.DatabaseOptions{
		Host:         env.GetValue("database host"),
		Port:         env.GetValue("database port"),
		DatabaseName: env.GetValue("database name"),
		Username:     env.GetValue("database username"),
		Password:     env.GetValue("database password"),
		SSLMode:      env.GetValue("database ssl mode"),
	}

	log.Info("🔌 Starting database connection process")

	// Warn about beginning validation
	log.Warning("⚠️ Validating database options")

	// Validate required fields
	if err := validateDatabaseOptions(databaseOptions); err != nil {
		log.Error(fmt.Sprintf("❌ Invalid database configuration: %v", err))
		return nil, err
	}

	// Open a connection to the PostgreSQL database using the provided URI.
	log.Info("📡 Opening connection to PostgreSQL database")
	db, err := sql.Open("postgres", getURI(databaseOptions))
	if err != nil {
		log.Error(fmt.Sprintf("❌ Failed to open database connection: %v", err))
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// Configure connection pool settings
	log.Info("⚙️ Configuring database connection pool")
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(500)
	db.SetConnMaxLifetime(2 * time.Hour)
	db.SetConnMaxIdleTime(15 * time.Minute)

	// Verify connectivity
	log.Info("🔎 Verifying database connectivity with ping")
	if err := db.Ping(); err != nil {
		log.Error(fmt.Sprintf("❌ Failed to ping database: %v", err))
		return nil, fmt.Errorf("unable to connect to the database: %w", err)
	}

	log.Success(fmt.Sprintf("✅ Successfully connected to database: %s", databaseOptions.DatabaseName))
	return db, nil
}
