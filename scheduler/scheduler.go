package scheduler

import (
	"context"   // context provides support for cancellation and timeouts.
	"fmt"       // fmt provides formatting and printing functions.
	"os"        // os provides file system operations and signal handling.
	"os/signal" // signal provides system signal handling.
	"runtime"   // runtime provides access to system resources.
	"sync"      // sync provides synchronization primitives.
	"syscall"   // syscall provides system call constants.
	"time"      // time provides functionality for handling intervals and sleeping.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// SchedulerConfig holds configuration parameters for the scheduler.
// This struct centralizes all scheduler settings for better maintainability.
type SchedulerConfig struct {
	Interval               time.Duration // Interval specifies the duration between function executions
	RunInstant             bool          // RunInstant specifies whether to run the function immediately before the first interval
	EnableGracefulShutdown bool          // EnableGracefulShutdown specifies whether to handle OS signals for graceful shutdown
	MaxPanicRecovery       int           // MaxPanicRecovery specifies maximum consecutive panics before stopping (0 = unlimited)
}

// LoadConfig loads scheduler configuration with defaults.
// Returns a SchedulerConfig struct with default values.
func LoadConfig(interval time.Duration, runInstant bool) SchedulerConfig {
	return SchedulerConfig{
		Interval:               interval,
		RunInstant:             runInstant,
		EnableGracefulShutdown: true,
		MaxPanicRecovery:       3, // Allow 3 consecutive panics before stopping
	}
}

// validateInterval validates that the interval is a positive duration.
// Returns an error if the interval is zero or negative.
func validateInterval(interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("interval must be positive, got: %v", interval)
	}
	return nil
}

// runWithRecovery executes a function with panic recovery and logging.
// Returns true if the function completed successfully, false if it panicked.
func runWithRecovery(functionToRun func(), operationName string) (success bool) {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic with detailed information
			log.Error(fmt.Sprintf("🚨 PANIC in %s: %v", operationName, r))

			// Capture stack trace for debugging
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			log.Warning(fmt.Sprintf("Stack trace: %s", string(buf[:n])))

			success = false
		}
	}()

	// Execute the function
	functionToRun()
	return true
}

// SchedulerState holds the current state of the scheduler for monitoring.
type SchedulerState struct {
	mu             sync.RWMutex
	StartTime      time.Time // StartTime records when the scheduler started
	ExecutionCount int64     // ExecutionCount tracks total successful executions
	PanicCount     int64     // PanicCount tracks total panic recoveries
	LastExecution  time.Time // LastExecution records when the function was last run
	LastError      string    // LastError stores the last error message
	IsRunning      bool      // IsRunning indicates if the scheduler is active
}

// NewSchedulerState creates and initializes a new SchedulerState.
func NewSchedulerState() *SchedulerState {
	return &SchedulerState{
		StartTime: time.Now(),
		IsRunning: true,
	}
}

// RecordExecution records a successful function execution.
func (s *SchedulerState) RecordExecution() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ExecutionCount++
	s.LastExecution = time.Now()
}

// RecordPanic records a panic occurrence.
func (s *SchedulerState) RecordPanic(err interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PanicCount++
	s.LastError = fmt.Sprintf("%v", err)
}

// Stop marks the scheduler as stopped.
func (s *SchedulerState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsRunning = false
}

// GetStatus returns the current scheduler status for monitoring.
func (s *SchedulerState) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"is_running":      s.IsRunning,
		"start_time":      s.StartTime,
		"execution_count": s.ExecutionCount,
		"panic_count":     s.PanicCount,
		"last_execution":  s.LastExecution,
		"last_error":      s.LastError,
		"uptime":          time.Since(s.StartTime).String(),
	}
}

// RunFunctionAtInterval schedules a function to run at regular intervals with graceful shutdown support.
// Executes the provided function repeatedly after the specified duration.
// Supports optional immediate execution before the first interval and graceful shutdown on OS signals.
//
// Parameters:
//   - functionToRun: The function to execute at each interval
//   - interval: The duration between function executions
//   - runInstant: If true, executes the function immediately before the first interval
//
// Example:
//
//	scheduler.RunFunctionAtInterval(myFunction, 5*time.Minute, true)
func RunFunctionAtInterval(functionToRun func(), interval time.Duration, runInstant bool) {
	// Validate the interval duration
	if err := validateInterval(interval); err != nil {
		log.Error(fmt.Sprintf("❌ Scheduler validation failed: %v", err))
		return
	}

	// Load configuration
	config := LoadConfig(interval, runInstant)

	// Initialize scheduler state for monitoring
	state := NewSchedulerState()

	// Set up context for graceful shutdown
	var ctx context.Context
	var cancel context.CancelFunc

	if config.EnableGracefulShutdown {
		// Create context that cancels on OS signals (SIGINT, SIGTERM)
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		log.Info("✅ Graceful shutdown enabled (responds to SIGINT/SIGTERM)")
	} else {
		// Use background context without signal handling
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}

	// Log the start of the scheduler with configuration details
	log.Info(fmt.Sprintf("⏰ Scheduler started: Function will run every %v. Run instantly: %v", interval, runInstant))
	log.Info(fmt.Sprintf("📊 Configuration - Graceful shutdown: %v, Max panic recovery: %d",
		config.EnableGracefulShutdown, config.MaxPanicRecovery))

	// Execute the function immediately if runInstant is true
	if runInstant {
		log.Info("🚀 Executing function immediately before first interval...")

		if success := runWithRecovery(functionToRun, "initial execution"); success {
			state.RecordExecution()
			log.Success("✅ Initial execution completed successfully.")
		} else {
			state.RecordPanic("panic during initial execution")
			log.Warning("⚠️  Initial execution encountered issues but scheduler continues...")
		}
	}

	// Create a ticker for the specified interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track consecutive panics for circuit breaker pattern
	consecutivePanics := 0

	// Main scheduler loop
	for {
		select {
		case <-ticker.C:
			// Execute the scheduled function with panic recovery
			log.Warning("⚡ Executing scheduled function...")

			if success := runWithRecovery(functionToRun, "scheduled execution"); success {
				state.RecordExecution()
				consecutivePanics = 0 // Reset panic counter on success
				log.Success("✅ Function execution completed successfully.")

				// Log periodic status every 10 executions for monitoring
				if state.ExecutionCount%10 == 0 {
					status := state.GetStatus()
					log.Info(fmt.Sprintf("📈 Scheduler status - Executions: %d, Panics: %d, Uptime: %v",
						status["execution_count"], status["panic_count"], status["uptime"]))
				}
			} else {
				state.RecordPanic("panic during scheduled execution")
				consecutivePanics++
				log.Warning(fmt.Sprintf("⚠️  Function execution encountered issues (consecutive panics: %d)", consecutivePanics))

				// Circuit breaker: stop scheduler after too many consecutive panics
				if config.MaxPanicRecovery > 0 && consecutivePanics >= config.MaxPanicRecovery {
					log.Error(fmt.Sprintf("❌ Too many consecutive panics (%d), stopping scheduler for safety", consecutivePanics))
					state.Stop()
					return
				}
			}

		case <-ctx.Done():
			// Handle graceful shutdown
			state.Stop()
			log.Info("🛑 Received shutdown signal, stopping scheduler gracefully...")

			// Log final statistics
			status := state.GetStatus()
			log.Info(fmt.Sprintf("📊 Final statistics - Total executions: %d, Total panics: %d, Total uptime: %v",
				status["execution_count"], status["panic_count"], status["uptime"]))

			log.Success("✅ Scheduler shutdown completed successfully")
			return
		}
	}
}
