package database

import (
	"context"      // context provides support for cancellation and timeouts.
	"database/sql" // sql provides database connectivity and query execution.
	"fmt"          // fmt provides formatting and printing functions.
	"net/url"      // url provides utilities for parsing database URIs.
	// strconv provides string conversion utilities.
	"strings" // strings provides utilities for string manipulation.
	"time"    // time provides functionality for handling connection timeouts.

	"github.com/hekimapro/utils/helpers"
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
	"github.com/hekimapro/utils/models"
	_ "github.com/lib/pq" // pq registers the PostgreSQL driver.
)

// DatabaseConfig holds configuration for database connection and connection pooling.
type DatabaseConfig struct {
	MaxIdleConns    int           // MaxIdleConns sets the maximum number of connections in the idle connection pool
	MaxOpenConns    int           // MaxOpenConns sets the maximum number of open connections to the database
	ConnMaxLifetime time.Duration // ConnMaxLifetime sets the maximum amount of time a connection may be reused
	ConnMaxIdleTime time.Duration // ConnMaxIdleTime sets the maximum amount of time a connection may be idle
	ConnectTimeout  time.Duration // ConnectTimeout sets the maximum time for establishing connection
	PingTimeout     time.Duration // PingTimeout sets the maximum time for ping operations
}

// LoadDatabaseConfig loads database configuration with defaults from environment variables.
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		MaxIdleConns:    helpers.GetENVIntValue("database maximum idle connections", 5),
		MaxOpenConns:    helpers.GetENVIntValue("database maximum open connections", 5),
		ConnMaxLifetime: time.Duration(helpers.GetENVIntValue("database connection maximum lifetime", 60)) * time.Minute,
		ConnMaxIdleTime: time.Duration(helpers.GetENVIntValue("database connection maximum idle time", 5)) * time.Minute,
		ConnectTimeout:  time.Duration(helpers.GetENVIntValue("database connect timeout", 30)) * time.Second,
		PingTimeout:     time.Duration(helpers.GetENVIntValue("database ping timeout", 10)) * time.Second,
	}
}

// getURI constructs the PostgreSQL connection URI from database options.
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
		missing = append(missing, "DATABASE_USERNAME")
	}
	if strings.TrimSpace(opts.Password) == "" {
		missing = append(missing, "DATABASE_PASSWORD")
	}
	if strings.TrimSpace(opts.Host) == "" {
		missing = append(missing, "DATABASE_HOST")
	}
	if strings.TrimSpace(opts.Port) == "" {
		missing = append(missing, "DATABASE_PORT")
	}
	if strings.TrimSpace(opts.DatabaseName) == "" {
		missing = append(missing, "DATABASE_NAME")
	}
	if strings.TrimSpace(opts.SSLMode) == "" {
		missing = append(missing, "DATABASE_SSL_MODE")
	}

	if len(missing) > 0 {
		return fmt.Errorf(".env file is missing required database option(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

// connectToDatabaseWithContext is the internal implementation with context support.
func connectToDatabaseWithContext(ctx context.Context) (*sql.DB, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "database connection cancelled before start")
	default:
		// Continue with connection
	}

	databaseOptions := models.DatabaseOptions{
		Host:         helpers.GetENVValue("database host"),
		Port:         helpers.GetENVValue("database port"),
		DatabaseName: helpers.GetENVValue("database name"),
		Username:     helpers.GetENVValue("database username"),
		Password:     helpers.GetENVValue("database password"),
		SSLMode:      helpers.GetENVValue("database ssl mode"),
	}

	log.Info("🔌 Starting database connection process")

	// Warn about beginning validation
	log.Warning("⚠️ Validating database options")

	// Validate required fields
	if err := validateDatabaseOptions(databaseOptions); err != nil {
		log.Error(fmt.Sprintf("❌ Invalid database configuration: %v", err))
		return nil, err
	}

	// Check context cancellation after validation
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "database connection cancelled after validation")
	default:
		// Continue with connection
	}

	// Load database configuration
	config := LoadDatabaseConfig()

	// Open a connection to the PostgreSQL database using the provided URI.
	log.Info("📡 Opening connection to PostgreSQL database")
	db, err := sql.Open("postgres", getURI(databaseOptions))
	if err != nil {
		log.Error(fmt.Sprintf("❌ Failed to open database connection: %v", err))
		return nil, helpers.WrapError(err, "unable to open database connection")
	}

	// Check context cancellation after opening connection
	select {
	case <-ctx.Done():
		db.Close()
		return nil, helpers.WrapError(ctx.Err(), "database connection cancelled after opening")
	default:
		// Continue with configuration
	}

	// Configure connection pool settings
	log.Info("⚙️ Configuring database connection pool")
	log.Info(fmt.Sprintf("📊 Connection pool settings - MaxIdle: %d, MaxOpen: %d, MaxLifetime: %v, MaxIdleTime: %v",
		config.MaxIdleConns, config.MaxOpenConns, config.ConnMaxLifetime, config.ConnMaxIdleTime))

	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Check context cancellation after pool configuration
	select {
	case <-ctx.Done():
		db.Close()
		return nil, helpers.WrapError(ctx.Err(), "database connection cancelled after pool configuration")
	default:
		// Continue with ping
	}

	// Verify connectivity with context timeout
	log.Info("🔎 Verifying database connectivity with ping")
	pingCtx, pingCancel := context.WithTimeout(ctx, config.PingTimeout)
	defer pingCancel()

	if err := db.PingContext(pingCtx); err != nil {
		log.Error(fmt.Sprintf("❌ Failed to ping database: %v", err))
		db.Close()
		return nil, helpers.WrapError(err, "unable to connect to the database")
	}

	// Check context cancellation after successful ping
	select {
	case <-ctx.Done():
		db.Close()
		return nil, helpers.WrapError(ctx.Err(), "database connection cancelled after successful ping")
	default:
		// Continue with success
	}

	log.Success(fmt.Sprintf("✅ Successfully connected to database: %s", databaseOptions.DatabaseName))
	log.Info(fmt.Sprintf("📈 Database connection pool configured - Idle: %d, Open: %d", config.MaxIdleConns, config.MaxOpenConns))

	return db, nil
}

// ConnectToDatabase establishes a connection to a PostgreSQL database.
// Configures connection pooling and verifies connectivity.
// Returns the database handle or an error if the connection fails.
func ConnectToDatabase() (*sql.DB, error) {
	// Create context with timeout for database connection
	config := LoadDatabaseConfig()
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	return connectToDatabaseWithContext(ctx)
}

// PingDatabase pings the database to verify connectivity with context support.
func PingDatabase(db *sql.DB) error {
	// Create context with timeout for ping operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pingDatabaseWithContext(ctx, db)
}

// pingDatabaseWithContext pings the database with context support.
func pingDatabaseWithContext(ctx context.Context, db *sql.DB) error {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "database ping cancelled before start")
	default:
		// Continue with ping
	}

	log.Info("🔍 Pinging database to verify connectivity")
	if err := db.PingContext(ctx); err != nil {
		log.Error(fmt.Sprintf("❌ Database ping failed: %v", err))
		return helpers.WrapError(err, "database ping failed")
	}

	log.Success("✅ Database ping successful")
	return nil
}

// CloseDatabase closes the database connection with context support.
func CloseDatabase(db *sql.DB) error {
	// Create context with timeout for close operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return closeDatabaseWithContext(ctx, db)
}

// closeDatabaseWithContext closes the database connection with context support.
func closeDatabaseWithContext(ctx context.Context, db *sql.DB) error {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "database close cancelled before start")
	default:
		// Continue with close
	}

	log.Info("🔌 Closing database connection")

	// Use a channel to handle the close operation with context
	closeDone := make(chan error, 1)
	go func() {
		closeDone <- db.Close()
	}()

	// Wait for either the close to complete or context cancellation
	select {
	case <-ctx.Done():
		log.Warning("⚠️ Database close operation cancelled or timed out")
		return helpers.WrapError(ctx.Err(), "database close cancelled")
	case err := <-closeDone:
		if err != nil {
			log.Error(fmt.Sprintf("❌ Failed to close database connection: %v", err))
			return helpers.WrapError(err, "failed to close database connection")
		}
		log.Success("✅ Database connection closed successfully")
		return nil
	}
}

// GetDatabaseStats returns database connection pool statistics.
func GetDatabaseStats(db *sql.DB) sql.DBStats {
	return db.Stats()
}

// PrintDatabaseStats logs database connection pool statistics.
func PrintDatabaseStats(db *sql.DB) {
	stats := GetDatabaseStats(db)

	log.Info("📊 Database Connection Pool Statistics:")
	log.Info(fmt.Sprintf("   Open Connections: %d", stats.OpenConnections))
	log.Info(fmt.Sprintf("   In Use: %d", stats.InUse))
	log.Info(fmt.Sprintf("   Idle: %d", stats.Idle))
	log.Info(fmt.Sprintf("   Wait Count: %d", stats.WaitCount))
	log.Info(fmt.Sprintf("   Wait Duration: %v", stats.WaitDuration))
	log.Info(fmt.Sprintf("   Max Idle Closed: %d", stats.MaxIdleClosed))
	log.Info(fmt.Sprintf("   Max Lifetime Closed: %d", stats.MaxLifetimeClosed))
}

// IsDatabaseConnected checks if the database is connected and responsive.
func IsDatabaseConnected(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return isDatabaseConnectedWithContext(ctx, db)
}

// isDatabaseConnectedWithContext checks if the database is connected with context support.
func isDatabaseConnectedWithContext(ctx context.Context, db *sql.DB) bool {
	if db == nil {
		return false
	}

	if err := db.PingContext(ctx); err != nil {
		return false
	}

	return true
}

// QueryRowWithContext is a convenience function for querying a single row with context.
func QueryRowWithContext(ctx context.Context, db *sql.DB, query string, args ...interface{}) *sql.Row {
	return db.QueryRowContext(ctx, query, args...)
}

// QueryWithContext is a convenience function for querying multiple rows with context.
func QueryWithContext(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

// ExecWithContext is a convenience function for executing queries with context.
func ExecWithContext(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

// GetDatabaseVersion returns the PostgreSQL server version.
func GetDatabaseVersion(db *sql.DB) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return getDatabaseVersionWithContext(ctx, db)
}

// GetDatabaseVersionWithContext returns the PostgreSQL server version with context support.
func getDatabaseVersionWithContext(ctx context.Context, db *sql.DB) (string, error) {
	var version string
	err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return "", helpers.WrapError(err, "failed to get database version")
	}
	return version, nil
}