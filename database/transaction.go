package database

import (
	"context"      // context provides support for cancellation and timeouts.
	"database/sql" // sql provides database connectivity and transaction management.
	"errors"
	"fmt"  // fmt provides formatting and printing functions.
	"time" // time provides functionality for tracking transaction duration.

	"github.com/hekimapro/utils/helpers" // helpers provides utility functions.
	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
)

// TransactionFunction defines the signature for the transactional operation.
// It accepts a transaction and returns an error if the operation fails.
type TransactionFunction func(transaction *sql.Tx) error

// Transaction executes a database transaction with robust panic handling,
// consistent timestamps, and enhanced logging for debugging.
// It returns an error if the transaction fails or panics.
func Transaction(database *sql.DB, operation TransactionFunction) (err error) {
	// Create context with timeout for transaction operation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return transactionWithContext(ctx, database, operation)
}

// transactionWithContext is the internal implementation with context support.
func transactionWithContext(ctx context.Context, database *sql.DB, operation TransactionFunction) (err error) {
	const timestampFormat = "2006-01-02 15:04:05.000" // ISO 8601-like format for timestamps.

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "transaction cancelled before start")
	default:
		// Continue with transaction
	}

	// Record the start time of the transaction.
	startTime := time.Now()
	// Log the start of the transaction with timestamp.
	log.Info(fmt.Sprintf("🔄 Beginning DB transaction at %s", startTime.Format(timestampFormat)))
	// Log that the transactional operation is being executed.
	log.Info("🛠️  Executing transactional operation...")

	// Begin the database transaction with context support.
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		// Log and return an error if starting the transaction fails.
		log.Error(fmt.Sprintf("❌ Failed to begin transaction: %s", err.Error()))
		return helpers.WrapError(err, "failed to start transaction")
	}

	// Defer transaction cleanup (commit or rollback) and panic/error handling.
	defer func() {
		// Check context cancellation in defer
		select {
		case <-ctx.Done():
			log.Warning("⚠️ Transaction context cancelled during cleanup")
		default:
			// Continue with cleanup
		}

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
				err = helpers.WrapError(rollbackErr, "rollback failed after panic")
			} else {
				// Log successful rollback due to panic.
				log.Warning("⚠️  Transaction rolled back due to panic")
				err = helpers.WrapError(fmt.Errorf("%v", recovered), "transaction panicked")
			}
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
				err = helpers.WrapError(rollbackErr, "rollback failed")
			} else {
				// Log successful rollback due to error.
				log.Warning("⚠️  Transaction rolled back due to error")
			}
			return
		}

		// Check context cancellation before commit
		select {
		case <-ctx.Done():
			log.Warning("⚠️ Transaction context cancelled before commit, rolling back")
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("❌ Rollback failed after context cancellation: %s", rollbackErr.Error()))
				err = helpers.WrapError(rollbackErr, "rollback failed after context cancellation")
			} else {
				err = helpers.WrapError(ctx.Err(), "transaction cancelled before commit")
			}
			return
		default:
			// Continue with commit
		}

		// Attempt to commit the transaction if no errors occurred.
		log.Info("📝 Committing transaction...")
		if commitErr := transaction.Commit(); commitErr != nil {
			// Log and set error if commit fails.
			log.Error(fmt.Sprintf("❌ Commit failed: %s", commitErr.Error()))
			err = helpers.WrapError(commitErr, "failed to commit transaction")
		} else {
			// Log successful transaction commit.
			log.Success("✅ Transaction committed successfully")
		}
	}()

	// Check context cancellation before executing operation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "transaction cancelled before operation execution")
	default:
		// Continue with operation
	}

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

// TransactionWithRetry executes a database transaction with retry logic for transient errors.
func TransactionWithRetry(database *sql.DB, operation TransactionFunction, maxRetries int) error {
	// Create context with timeout for retry operation
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	return transactionWithRetryAndContext(ctx, database, operation, maxRetries)
}

// transactionWithRetryAndContext is the internal implementation with context support and retry logic.
func transactionWithRetryAndContext(ctx context.Context, database *sql.DB, operation TransactionFunction, maxRetries int) error {
	// Validate maxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	var lastError error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return helpers.WrapError(ctx.Err(), "transaction with retry cancelled")
		default:
			// Continue with attempt
		}

		// Log retry attempt if not the first attempt
		if attempt > 0 {
			log.Warning(fmt.Sprintf("🔄 Transaction retry attempt %d/%d", attempt, maxRetries))
			// Exponential backoff
			backoffDuration := time.Duration(attempt*attempt) * time.Second
			if backoffDuration > 10*time.Second {
				backoffDuration = 10 * time.Second
			}

			select {
			case <-ctx.Done():
				return helpers.WrapError(ctx.Err(), "transaction retry cancelled during backoff")
			case <-time.After(backoffDuration):
				// Continue with next attempt
			}
		}

		// Execute the transaction
		err := transactionWithContext(ctx, database, operation)
		if err == nil {
			// Success
			if attempt > 0 {
				log.Success(fmt.Sprintf("✅ Transaction succeeded on attempt %d", attempt+1))
			}
			return nil
		}

		lastError = err

		// Check if error is retryable
		if !isRetryableTransactionError(err) {
			log.Warning("⚠️ Non-retryable transaction error, not retrying")
			return err
		}

		// Log that we will retry
		if attempt < maxRetries {
			log.Warning(fmt.Sprintf("⚠️ Retryable transaction error, will retry: %v", err))
		}
	}

	log.Error(fmt.Sprintf("❌ Transaction failed after %d attempts: %v", maxRetries+1, lastError))
	return helpers.WrapError(lastError, "transaction failed after maximum retries")
}

// isRetryableTransactionError checks if a transaction error is retryable.
func isRetryableTransactionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for retryable database errors using our error analysis package
	dbError := AnalyzeDatabaseError(err)
	if dbError != nil {
		retryableErrors := map[string]bool{
			"lock_timeout":        true,
			"too_many_connections": true,
			"query_cancelled":     true,
		}
		return retryableErrors[dbError.ErrorType]
	}

	// Default: don't retry on unknown errors
	return false
}

// TransactionWithIsolation executes a database transaction with a specific isolation level.
func TransactionWithIsolation(database *sql.DB, operation TransactionFunction, isolationLevel sql.IsolationLevel) error {
	// Create context with timeout for isolation transaction
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return transactionWithIsolationAndContext(ctx, database, operation, isolationLevel)
}

// transactionWithIsolationAndContext is the internal implementation with context support and isolation level.
func transactionWithIsolationAndContext(ctx context.Context, database *sql.DB, operation TransactionFunction, isolationLevel sql.IsolationLevel) error {
	const timestampFormat = "2006-01-02 15:04:05.000"

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "isolation transaction cancelled before start")
	default:
		// Continue with transaction
	}

	// Record the start time of the transaction.
	startTime := time.Now()

	// Log the isolation level
	isolationName := getIsolationLevelName(isolationLevel)
	log.Info(fmt.Sprintf("🔄 Beginning DB transaction with isolation level '%s' at %s",
		isolationName, startTime.Format(timestampFormat)))

	// Begin the database transaction with specific isolation level.
	transaction, err := database.BeginTx(ctx, &sql.TxOptions{
		Isolation: isolationLevel,
		ReadOnly:  false,
	})
	if err != nil {
		log.Error(fmt.Sprintf("❌ Failed to begin transaction with isolation level '%s': %s", isolationName, err.Error()))
		return helpers.WrapError(err, "failed to start transaction with isolation level")
	}

	// Defer transaction cleanup
	defer func() {
		// Record the end time and calculate transaction duration.
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		log.Info(fmt.Sprintf("🕒 Isolation transaction ended at %s (duration: %s)", endTime.Format(timestampFormat), duration))

		// Handle panic
		if recovered := recover(); recovered != nil {
			log.Error(fmt.Sprintf("💥 Panic during isolation transaction: %v", recovered))
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("❌ Rollback failed after panic: %s", rollbackErr.Error()))
			} else {
				log.Warning("⚠️  Isolation transaction rolled back due to panic")
			}
			err = helpers.WrapError(fmt.Errorf("%v", recovered), "isolation transaction panicked")
			return
		}

		// Handle error
		if err != nil {
			log.Error(fmt.Sprintf("❌ Isolation transaction operation error: %s", err.Error()))
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("❌ Rollback failed: %s", rollbackErr.Error()))
				err = helpers.WrapError(rollbackErr, "rollback failed")
			} else {
				log.Warning("⚠️  Isolation transaction rolled back due to error")
			}
			return
		}

		// Check context cancellation before commit
		select {
		case <-ctx.Done():
			log.Warning("⚠️ Isolation transaction context cancelled before commit, rolling back")
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				log.Error(fmt.Sprintf("❌ Rollback failed after context cancellation: %s", rollbackErr.Error()))
				err = helpers.WrapError(rollbackErr, "rollback failed after context cancellation")
			} else {
				err = helpers.WrapError(ctx.Err(), "isolation transaction cancelled before commit")
			}
			return
		default:
			// Continue with commit
		}

		// Commit
		log.Info("📝 Committing isolation transaction...")
		if commitErr := transaction.Commit(); commitErr != nil {
			log.Error(fmt.Sprintf("❌ Commit failed: %s", commitErr.Error()))
			err = helpers.WrapError(commitErr, "failed to commit isolation transaction")
		} else {
			log.Success("✅ Isolation transaction committed successfully")
		}
	}()

	// Check context cancellation before executing operation
	select {
	case <-ctx.Done():
		return helpers.WrapError(ctx.Err(), "isolation transaction cancelled before operation execution")
	default:
		// Continue with operation
	}

	// Execute the operation
	err = operation(transaction)
	if err != nil {
		log.Error(fmt.Sprintf("⚠️  Isolation transaction operation failed: %s", err.Error()))
	} else {
		log.Info("✔️  Isolation transaction operation completed successfully")
	}

	return err
}

// getIsolationLevelName returns a human-readable name for the isolation level.
func getIsolationLevelName(level sql.IsolationLevel) string {
	switch level {
	case sql.LevelDefault:
		return "Default"
	case sql.LevelReadUncommitted:
		return "Read Uncommitted"
	case sql.LevelReadCommitted:
		return "Read Committed"
	case sql.LevelRepeatableRead:
		return "Repeatable Read"
	case sql.LevelSnapshot:
		return "Snapshot"
	case sql.LevelSerializable:
		return "Serializable"
	case sql.LevelLinearizable:
		return "Linearizable"
	default:
		return "Unknown"
	}
}

// TransactionMetrics holds metrics about transaction execution.
type TransactionMetrics struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Success      bool
	ErrorType    string
	RetryCount   int
	IsolationLevel string
}

// TransactionWithMetrics executes a transaction and returns detailed metrics.
func TransactionWithMetrics(database *sql.DB, operation TransactionFunction) (*TransactionMetrics, error) {
	// Create context with timeout for metrics transaction
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return transactionWithMetricsAndContext(ctx, database, operation)
}

// transactionWithMetricsAndContext is the internal implementation with context support and metrics.
func transactionWithMetricsAndContext(ctx context.Context, database *sql.DB, operation TransactionFunction) (*TransactionMetrics, error) {
	metrics := &TransactionMetrics{
		StartTime: time.Now(),
		IsolationLevel: "Default",
	}

	err := transactionWithContext(ctx, database, operation)

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Success = (err == nil)

	if err != nil {
		dbError := AnalyzeDatabaseError(err)
		if dbError != nil {
			metrics.ErrorType = dbError.ErrorType
		} else {
			metrics.ErrorType = "unknown"
		}
	}

	return metrics, err
}