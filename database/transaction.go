package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hekimapro/utils/log"
	"github.com/lib/pq"
)

// TransactionFunction defines the signature for the transactional operation.
type TransactionFunction func(transaction *sql.Tx) error

// Transaction executes a database transaction with retry logic for PostgreSQL,
// improved panic handling, simplified timestamps, and detailed logging.
func Transaction(database *sql.DB, operation TransactionFunction) (err error) {
	// Constants for retry logic and timestamp format
	const (
		maxRetries      = 5                         // Maximum number of retry attempts for transient errors
		baseDelay       = 500                       // Base delay in milliseconds for exponential backoff
		timestampFormat = "2006-01-02 15:04:05.000" // ISO 8601-like format for consistent timestamps
	)

	// Loop to handle retries for transient errors
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Record the start time for logging transaction duration
		startTime := time.Now()
		log.Info(fmt.Sprintf("🚀 Starting database transaction (attempt %d) at %s", attempt, startTime.Format(timestampFormat)))

		// Begin a new transaction
		transaction, err := database.Begin()
		if err != nil {
			log.Error(fmt.Sprintf("❌ Failed to begin transaction: %v", err.Error()))
			if attempt == maxRetries {
				// If maximum retries are exhausted, return the error wrapped with context
				return fmt.Errorf("failed to begin transaction after %d attempts: %w", maxRetries, err)
			}
			// Apply exponential backoff before retrying: baseDelay * 2^(attempt-1) milliseconds
			time.Sleep(time.Duration(baseDelay*(1<<(attempt-1))) * time.Millisecond)
			continue
		}

		// Flag to determine if a retry is needed after an error
		shouldRetry := false

		// Deferred block to handle cleanup, logging, and error/panic handling
		defer func() {
			// Calculate and log transaction duration and end time
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			log.Info(fmt.Sprintf("🕓 Transaction duration: %s", duration))
			log.Info(fmt.Sprintf("📍 Transaction ended at %s", endTime.Format(timestampFormat)))

			// Handle panics to prevent application crashes
			if recoveredPanic := recover(); recoveredPanic != nil {
				log.Error(fmt.Sprintf("🔥 Panic occurred during transaction: %v", recoveredPanic))
				// Attempt to rollback the transaction
				if rollbackError := transaction.Rollback(); rollbackError != nil {
					log.Error(fmt.Sprintf("⚠️ Failed to rollback after panic: %v", rollbackError.Error()))
				}
				// Convert panic to an error instead of rethrowing
				err = fmt.Errorf("transaction panicked: %v", recoveredPanic)
			} else if err != nil {
				// Handle errors from the operation or commit
				log.Error(fmt.Sprintf("❗ Transaction error: %v", err.Error()))
				log.Info("↩️ Rolling back transaction due to error.")
				// Attempt to rollback the transaction
				if rollbackError := transaction.Rollback(); rollbackError != nil {
					log.Error(fmt.Sprintf("⚠️ Failed to rollback: %v", rollbackError.Error()))
					// Wrap rollback error with original error for better context
					err = fmt.Errorf("failed to rollback: %w; original error: %v", rollbackError, err)
				}
				// Check if the error is retryable for PostgreSQL
				if attempt < maxRetries && isRetryableError(err) {
					shouldRetry = true
				}
			} else {
				// Attempt to commit the transaction if no errors occurred
				log.Info("✅ Attempting to commit transaction...")
				if commitError := transaction.Commit(); commitError != nil {
					log.Error(fmt.Sprintf("❌ Failed to commit transaction: %v", commitError.Error()))
					err = fmt.Errorf("failed to commit transaction: %w", commitError)
					// Check if the commit error is retryable
					if attempt < maxRetries && isRetryableError(err) {
						shouldRetry = true
					}
				} else {
					log.Info("✅ Transaction committed successfully.")
				}
			}
		}()

		// Execute the user-provided transactional operation
		log.Info("🔧 Executing transactional operation...")
		err = operation(transaction)
		if err != nil {
			log.Error(fmt.Sprintf("❌ Transactional operation failed: %v", err.Error()))
		} else {
			log.Info("✅ Transactional operation completed successfully.")
		}

		// Exit the retry loop if no retry is needed
		if !shouldRetry {
			break
		}
		// Apply exponential backoff before retrying
		time.Sleep(time.Duration(baseDelay*(1<<(attempt-1))) * time.Millisecond)
	}

	// Return the final error (nil if successful)
	return err
}

// isRetryableError determines if an error is retryable based on PostgreSQL error codes.
// It checks for transient errors that may resolve on retry, such as deadlocks or connection issues.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Convert error to pq.Error to access PostgreSQL-specific error codes
	if pqErr, ok := err.(*pq.Error); ok {
		// PostgreSQL error codes for retryable conditions (Class 40, 53, 08, etc.)
		// Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html
		switch pqErr.Code {
		case "40001": // serialization_failure (e.g., transaction conflicts)
			return true
		case "40P01": // deadlock_detected
			return true
		case "08000": // connection_exception
			return true
		case "08003": // connection_does_not_exist
			return true
		case "08006": // connection_failure
			return true
		case "53xxx": // Class 53: Insufficient Resources (e.g., too_many_connections)
			if strings.HasPrefix(string(pqErr.Code), "53") {
				return true
			}
		case "57P03": // cannot_connect_now (database in recovery mode)
			return true
		case "55P03": // lock_not_available (lock contention)
			return true
		}
	}

	// Additional heuristic: check for common transient error messages
	// This is a fallback for cases where the error code is not specific
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "deadlock") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "lock") {
		return true
	}

	return false
}
