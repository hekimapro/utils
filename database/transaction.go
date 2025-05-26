package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hekimapro/utils/log"
)

// TransactionFunction defines the signature for the transactional operation.
type TransactionFunction func(transaction *sql.Tx) error

// Transaction executes a database transaction with robust panic handling,
// consistent timestamps, and enhanced logging for debugging.
func Transaction(database *sql.DB, operation TransactionFunction) (err error) {
	// Constants for timestamp format
	const (
		timestampFormat = "2006-01-02 15:04:05.000" // ISO 8601-like format for consistent timestamps
	)

	// Record start time for transaction duration
	startTime := time.Now()
	log.Info(fmt.Sprintf("[DB] Starting transaction at %s", startTime.Format(timestampFormat)))

	// Begin a new transaction
	transaction, err := database.Begin()
	if err != nil {
		log.Error(fmt.Sprintf("[DB] Failed to start transaction: %s", err.Error()))
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Deferred block for cleanup, logging, and error/panic handling
	defer func() {
		// Log transaction duration and end time
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		log.Info(fmt.Sprintf("[DB] Transaction ended at %s (duration: %s)", endTime.Format(timestampFormat), duration))

		// Handle panics to prevent crashes
		if recoveredPanic := recover(); recoveredPanic != nil {
			log.Error(fmt.Sprintf("[DB] Panic in transaction: %v", recoveredPanic))
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("[DB] Rollback failed after panic: %s", rollbackErr.Error()))
			} else {
				log.Info("[DB] Rolled back transaction due to panic")
			}
			err = fmt.Errorf("transaction panicked: %v", recoveredPanic)
			return
		}

		// Handle errors from operation or commit
		if err != nil {
			log.Error(fmt.Sprintf("[DB] Transaction failed: %s", err.Error()))
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("[DB] Rollback failed: %s", rollbackErr.Error()))
				err = fmt.Errorf("rollback failed: %w; original error: %v", rollbackErr, err)
			} else {
				log.Info("[DB] Rolled back transaction due to error")
			}
			return
		}

		// Attempt to commit the transaction
		log.Info("[DB] Committing transaction")
		if commitErr := transaction.Commit(); commitErr != nil {
			log.Error(fmt.Sprintf("[DB] Commit failed: %s", commitErr.Error()))
			err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		} else {
			log.Info("[DB] Transaction committed successfully")
		}
	}()

	// Execute the transactional operation
	log.Info("[DB] Executing operation")
	err = operation(transaction)
	if err != nil {
		log.Error(fmt.Sprintf("[DB] Operation failed: %s", err.Error()))
	} else {
		log.Info("[DB] Operation completed successfully")
	}

	return err
}