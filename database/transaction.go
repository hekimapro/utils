package database

import (
	"database/sql" // sql provides database connectivity and transaction management.
	"fmt"          // fmt provides formatting and printing functions.
	"time"         // time provides functionality for tracking transaction duration.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// TransactionFunction defines the signature for the transactional operation.
// It accepts a transaction and returns an error if the operation fails.
type TransactionFunction func(transaction *sql.Tx) error

// Transaction executes a database transaction with robust panic handling,
// consistent timestamps, and enhanced logging for debugging.
// It returns an error if the transaction fails or panics.
func Transaction(database *sql.DB, operation TransactionFunction) (err error) {
	const timestampFormat = "2006-01-02 15:04:05.000" // ISO 8601-like format for timestamps.

	// Record the start time of the transaction.
	startTime := time.Now()
	// Log the start of the transaction with timestamp.
	log.Info(fmt.Sprintf("🔄 Beginning DB transaction at %s", startTime.Format(timestampFormat)))
	// Log that the transactional operation is being executed.
	log.Info("🛠️  Executing transactional operation...")

	// Begin the database transaction.
	transaction, err := database.Begin()
	if err != nil {
		// Log and return an error if starting the transaction fails.
		log.Error(fmt.Sprintf("❌ Failed to begin transaction: %s", err.Error()))
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Defer transaction cleanup (commit or rollback) and panic/error handling.
	defer func() {
		// Record the end time and calculate transaction duration.
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		// Log transaction completion with timestamp and duration.
		log.Info(fmt.Sprintf("🕒 Transaction ended at %s (duration: %s)", endTime.Format(timestampFormat), duration))

		// Handle any panic that occurred during the transaction.
		if recovered := recover(); recovered != nil {
			// Log the panic details.
			log.Error(fmt.Sprintf("💥 Panic during transaction: %v", recovered))
			// Attempt to rollback the transaction.
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				// Log if rollback fails after panic.
				log.Error(fmt.Sprintf("❌ Rollback failed after panic: %s", rollbackErr.Error()))
			} else {
				// Log successful rollback due to panic.
				log.Warning("⚠️  Transaction rolled back due to panic")
			}
			// Set the error to indicate a panic occurred.
			err = fmt.Errorf("transaction panicked: %v", recovered)
			return
		}

		// Handle any error from the transaction operation.
		if err != nil {
			// Log the operation error.
			log.Error(fmt.Sprintf("❌ Transaction operation error: %s", err.Error()))
			// Attempt to rollback the transaction.
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				// Log if rollback fails and combine errors.
				log.Error(fmt.Sprintf("❌ Rollback failed: %s", rollbackErr.Error()))
				err = fmt.Errorf("rollback failed: %w; original error: %v", rollbackErr, err)
			} else {
				// Log successful rollback due to error.
				log.Warning("⚠️  Transaction rolled back due to error")
			}
			return
		}

		// Attempt to commit the transaction if no errors occurred.
		log.Info("📝 Committing transaction...")
		if commitErr := transaction.Commit(); commitErr != nil {
			// Log and set error if commit fails.
			log.Error(fmt.Sprintf("❌ Commit failed: %s", commitErr.Error()))
			err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		} else {
			// Log successful transaction commit.
			log.Success("✅ Transaction committed successfully")
		}
	}()

	// Execute the provided transactional operation.
	err = operation(transaction)
	if err != nil {
		// Log if the operation fails.
		log.Error(fmt.Sprintf("⚠️  Transaction operation failed: %s", err.Error()))
	} else {
		// Log successful operation execution.
		log.Info("✔️  Transaction operation completed successfully")
	}

	return err
}