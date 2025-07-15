package scheduler

import (
	"fmt"  // fmt provides formatting and printing functions.
	"time" // time provides functionality for handling intervals and sleeping.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// RunFunctionAtInterval schedules a function to run at regular intervals.
// Executes the provided function repeatedly after the specified duration.
// Supports optional immediate execution before the first interval.
func RunFunctionAtInterval(functionToRun func(), interval time.Duration, runInstant bool) {
	// Log the start of the scheduler with interval and immediate execution details.
	log.Info(fmt.Sprintf("⏰ Scheduler started: Function will run every %v. Run instantly: %v", interval, runInstant))

	// Execute the function immediately if runInstant is true.
	if runInstant {
		log.Info("🚀 Executing function immediately before first interval...")
		functionToRun()
		// Log successful initial execution.
		log.Success("✅ Initial execution completed. Waiting for next interval...")
	}

	// Run the function indefinitely at the specified interval.
	for {
		// Log the wait for the next interval.
		log.Info(fmt.Sprintf("⌛ Waiting for the next interval: %v...", interval))
		// Pause execution until the interval elapses.
		time.Sleep(interval)

		// Log the start of the scheduled function execution.
		log.Warning("⚡ Executing scheduled function...")
		functionToRun()
		// Log successful function execution.
		log.Success("✅ Function execution completed.")
	}
}