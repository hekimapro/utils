package database

import (
	"context" // context provides support for cancellation and timeouts.
	"errors"  // errors provides error creation utilities.
	"fmt"     // fmt provides formatting and printing functions.
	"strings" // strings provides utilities for string manipulation.
	"time"    // time provides functionality for timeouts and durations.

	// helpers provides utility functions.
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
	"github.com/lib/pq"              // pq provides PostgreSQL driver error handling.
)

// DatabaseError represents a structured database error with context.
type DatabaseError struct {
	OriginalError error  // OriginalError is the underlying database error
	ErrorType     string // ErrorType categorizes the error (duplicate, foreign_key, etc.)
	Constraint    string // Constraint is the database constraint that was violated
	Message       string // Message is a user-friendly error message
	Table         string // Table is the database table where the error occurred
	Column        string // Column is the database column where the error occurred
}

// Error returns the string representation of the database error.
func (e *DatabaseError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.OriginalError.Error()
}

// Unwrap returns the original error for error wrapping compatibility.
func (e *DatabaseError) Unwrap() error {
	return e.OriginalError
}

// IsDuplicateError checks if the error is a Postgres duplicate entry error (unique_violation).
// Returns the constraint/field name if duplicate, otherwise nil.
func IsDuplicateError(err error) *string {
	// Create context with timeout for error analysis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return isDuplicateErrorWithContext(ctx, err)
}

// isDuplicateErrorWithContext is the internal implementation with context support.
func isDuplicateErrorWithContext(ctx context.Context, err error) *string {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		log.Warning("⚠️ Duplicate error check cancelled")
		return nil
	default:
		// Continue with error analysis
	}

	// For github.com/lib/pq
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			label := extractColumnLabel(pqErr.Constraint) + " already exists"
			log.Warning("⚠️ Duplicate entry detected: " + label)
			return &label
		}
	}

	return nil
}

// ExtractColumnLabel converts "users_email_address_key" to "email address"
func extractColumnLabel(constraint string) string {
	constraint = strings.TrimSuffix(constraint, "_key")
	parts := strings.Split(constraint, "_")

	if len(parts) <= 1 {
		return constraint
	}

	// Remove the table name (assumed to be the first part)
	columnParts := parts[1:]
	return strings.Join(columnParts, " ")
}

// AnalyzeDatabaseError comprehensively analyzes database errors and returns structured information.
func AnalyzeDatabaseError(err error) *DatabaseError {
	// Create context with timeout for error analysis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return analyzeDatabaseErrorWithContext(ctx, err)
}

// analyzeDatabaseErrorWithContext is the internal implementation with context support.
func analyzeDatabaseErrorWithContext(ctx context.Context, err error) *DatabaseError {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		log.Warning("⚠️ Database error analysis cancelled")
		return &DatabaseError{
			OriginalError: err,
			ErrorType:     "analysis_cancelled",
			Message:       "Error analysis was cancelled",
		}
	default:
		// Continue with error analysis
	}

	if err == nil {
		return nil
	}

	log.Info("🔍 Analyzing database error")

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		dbError := &DatabaseError{
			OriginalError: err,
			Constraint:    pqErr.Constraint,
			Table:         pqErr.Table,
			Column:        pqErr.Column,
		}

		switch pqErr.Code {
		case "23505": // unique_violation
			dbError.ErrorType = "duplicate"
			dbError.Message = generateDuplicateErrorMessage(pqErr.Constraint, pqErr.Table)
			log.Warning("⚠️ Duplicate entry error: " + dbError.Message)

		case "23503": // foreign_key_violation
			dbError.ErrorType = "foreign_key"
			dbError.Message = generateForeignKeyErrorMessage(pqErr.Constraint, pqErr.Table)
			log.Warning("⚠️ Foreign key violation: " + dbError.Message)

		case "23502": // not_null_violation
			dbError.ErrorType = "not_null"
			dbError.Message = generateNotNullErrorMessage(pqErr.Column, pqErr.Table)
			log.Warning("⚠️ Not null violation: " + dbError.Message)

		case "23514": // check_violation
			dbError.ErrorType = "check_constraint"
			dbError.Message = generateCheckConstraintErrorMessage(pqErr.Constraint)
			log.Warning("⚠️ Check constraint violation: " + dbError.Message)

		case "42P01": // undefined_table
			dbError.ErrorType = "undefined_table"
			dbError.Message = fmt.Sprintf("Table '%s' does not exist", pqErr.Table)
			log.Error("❌ Undefined table: " + dbError.Message)

		case "42703": // undefined_column
			dbError.ErrorType = "undefined_column"
			dbError.Message = fmt.Sprintf("Column '%s' does not exist in table '%s'", pqErr.Column, pqErr.Table)
			log.Error("❌ Undefined column: " + dbError.Message)

		case "28000", "28P01": // invalid_authorization
			dbError.ErrorType = "authentication"
			dbError.Message = "Database authentication failed"
			log.Error("❌ Authentication failed")

		case "55P03": // lock_not_available
			dbError.ErrorType = "lock_timeout"
			dbError.Message = "Database lock timeout occurred"
			log.Warning("⚠️ Lock timeout occurred")

		case "53300": // too_many_connections
			dbError.ErrorType = "too_many_connections"
			dbError.Message = "Too many database connections"
			log.Error("❌ Too many database connections")

		case "57014": // query_canceled
			dbError.ErrorType = "query_cancelled"
			dbError.Message = "Database query was cancelled"
			log.Warning("⚠️ Query cancelled")

		default:
			dbError.ErrorType = "unknown"
			dbError.Message = fmt.Sprintf("Database error: %s", pqErr.Message)
			log.Error("❌ Unknown database error: " + pqErr.Message)
		}

		return dbError
	}

	// Handle non-PQ errors
	return &DatabaseError{
		OriginalError: err,
		ErrorType:     "generic",
		Message:       err.Error(),
	}
}

// generateDuplicateErrorMessage creates a user-friendly duplicate error message.
func generateDuplicateErrorMessage(constraint, table string) string {
	if constraint == "" {
		return "A record with these details already exists"
	}

	fieldName := extractColumnLabel(constraint)
	if fieldName != "" {
		return fmt.Sprintf("%s already exists", fieldName)
	}

	return "A duplicate record already exists"
}

// generateForeignKeyErrorMessage creates a user-friendly foreign key error message.
func generateForeignKeyErrorMessage(constraint, table string) string {
	if constraint == "" {
		return "The referenced record does not exist"
	}

	// Extract relationship from constraint name
	// Format: {table}_{column}_fkey or similar
	parts := strings.Split(constraint, "_")
	if len(parts) >= 2 {
		relatedTable := parts[0]
		return fmt.Sprintf("The referenced record in '%s' does not exist", relatedTable)
	}

	return "Referenced record does not exist"
}

// generateNotNullErrorMessage creates a user-friendly not null error message.
func generateNotNullErrorMessage(column, table string) string {
	if column == "" {
		return "Required field cannot be empty"
	}

	fieldName := strings.ReplaceAll(column, "_", " ")
	return fmt.Sprintf("%s is required and cannot be empty", fieldName)
}

// generateCheckConstraintErrorMessage creates a user-friendly check constraint error message.
func generateCheckConstraintErrorMessage(constraint string) string {
	if constraint == "" {
		return "The provided data does not meet the required constraints"
	}

	return fmt.Sprintf("Data validation failed for constraint '%s'", constraint)
}

// IsForeignKeyError checks if the error is a foreign key violation.
func IsForeignKeyError(err error) bool {
	dbError := AnalyzeDatabaseError(err)
	return dbError != nil && dbError.ErrorType == "foreign_key"
}

// IsNotNullError checks if the error is a not null violation.
func IsNotNullError(err error) bool {
	dbError := AnalyzeDatabaseError(err)
	return dbError != nil && dbError.ErrorType == "not_null"
}

// IsConnectionError checks if the error is related to database connection.
func IsConnectionError(err error) bool {
	dbError := AnalyzeDatabaseError(err)
	if dbError == nil {
		return false
	}

	// Check for connection-related error types
	connectionErrors := map[string]bool{
		"authentication":       true,
		"too_many_connections": true,
	}

	return connectionErrors[dbError.ErrorType] ||
		strings.Contains(strings.ToLower(err.Error()), "connection") ||
		strings.Contains(strings.ToLower(err.Error()), "connect")
}

// IsTimeoutError checks if the error is related to operation timeout.
func IsTimeoutError(err error) bool {
	dbError := AnalyzeDatabaseError(err)
	if dbError == nil {
		return false
	}

	return dbError.ErrorType == "query_cancelled" ||
		strings.Contains(strings.ToLower(err.Error()), "timeout") ||
		strings.Contains(strings.ToLower(err.Error()), "deadline")
}

// GetPostgresErrorCode returns the PostgreSQL error code if available.
func GetPostgresErrorCode(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code)
	}
	return ""
}

// IsPostgresError checks if the error is a PostgreSQL-specific error.
func IsPostgresError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr)
}

// FormatDatabaseError creates a user-friendly error message from a database error.
func FormatDatabaseError(err error) string {
	dbError := AnalyzeDatabaseError(err)
	if dbError == nil {
		return "An unexpected database error occurred"
	}

	if dbError.Message != "" {
		return dbError.Message
	}

	// Fallback to generic message
	switch dbError.ErrorType {
	case "duplicate":
		return "This record already exists"
	case "foreign_key":
		return "The referenced record does not exist"
	case "not_null":
		return "Required information is missing"
	case "authentication":
		return "Database connection failed"
	case "timeout":
		return "Database operation timed out"
	default:
		return "A database error occurred"
	}
}

// ShouldRetryError checks if the error is transient and the operation can be retried.
func ShouldRetryError(err error) bool {
	dbError := AnalyzeDatabaseError(err)
	if dbError == nil {
		return false
	}

	// Retry on connection issues, timeouts, and deadlocks
	retryableErrors := map[string]bool{
		"lock_timeout":         true,
		"query_cancelled":      true,
		"authentication":       false, // Don't retry auth errors
		"too_many_connections": true,  // Might be transient
		"duplicate":            false, // Don't retry duplicates
		"foreign_key":          false, // Don't retry FK violations
		"not_null":             false, // Don't retry constraint violations
	}

	if retryable, exists := retryableErrors[dbError.ErrorType]; exists {
		return retryable
	}

	// Default: retry on connection and timeout issues
	return IsConnectionError(err) || IsTimeoutError(err)
}

// GetErrorSeverity returns the severity level of the database error.
func GetErrorSeverity(err error) string {
	dbError := AnalyzeDatabaseError(err)
	if dbError == nil {
		return "unknown"
	}

	severityMap := map[string]string{
		"duplicate":            "warning",
		"foreign_key":          "error",
		"not_null":             "error",
		"check_constraint":     "error",
		"undefined_table":      "critical",
		"undefined_column":     "critical",
		"authentication":       "critical",
		"too_many_connections": "critical",
		"lock_timeout":         "warning",
		"query_cancelled":      "warning",
	}

	if severity, exists := severityMap[dbError.ErrorType]; exists {
		return severity
	}

	return "error"
}

// LogDatabaseError logs a database error with appropriate severity and context.
func LogDatabaseError(err error, operation string) {
	if err == nil {
		return
	}

	dbError := AnalyzeDatabaseError(err)
	severity := GetErrorSeverity(err)

	logMessage := fmt.Sprintf("Database error during %s: %s", operation, FormatDatabaseError(err))

	switch severity {
	case "warning":
		log.Warning("⚠️ " + logMessage)
	case "critical":
		log.Error("❌ " + logMessage)
	default:
		log.Error("❌ " + logMessage)
	}

	// Log additional context for debugging
	if dbError != nil && dbError.OriginalError != nil {
		log.Info(fmt.Sprintf("📋 Error details - Type: %s, Constraint: %s, Table: %s, Column: %s",
			dbError.ErrorType, dbError.Constraint, dbError.Table, dbError.Column))
	}
}
