package scheduler

import (
	"log"
	"time"
)

// RunFunctionAtInterval schedules a function to run at regular intervals
// Executes the provided function repeatedly after the specified duration
// Supports optional immediate execution before the first interval
func RunFunctionAtInterval(functionToRun func(), interval time.Duration, runInstant bool) {
	// Log the start of the scheduler with interval and immediate execution details
	log.Printf("Scheduler started: Function will run every %v. Run instantly: %v", interval, runInstant)

	// Conditionally execute the function immediately if runInstant is true
	if runInstant {
		// Log immediate execution start
		log.Printf("Executing function immediately before first interval...")
		functionToRun()
		// Log completion of immediate execution
		log.Printf("Initial execution completed. Waiting for interval...")
	}

	// Start an infinite loop to repeatedly execute the function
	for {
		// Log the wait for the next interval
		log.Printf("Waiting for the next interval of %v...", interval)
		// Pause execution for the specified interval
		time.Sleep(interval)

		// Log the start of function execution
		log.Printf("Executing function...")
		functionToRun()
		// Log completion of function execution
		log.Printf("Function execution completed.")
	}
}
