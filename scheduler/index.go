package scheduler

import (
	"log"
	"time"
)

// RunFunctionAtInterval receives a function and a time interval.
// It runs the provided function repeatedly after the given time interval.
func RunFunctionAtInterval(functionToRun func(), interval time.Duration) {
	// Start logging the function execution details
	log.Printf("Scheduler started: Function will run every %v.", interval)

	// Infinite loop to keep calling the function at the given interval
	for {
		// Log the start of the sleep interval
		log.Printf("Waiting for the next interval of %v...", interval)

		// Wait for the specified interval
		time.Sleep(interval)

		// Log before calling the function
		log.Printf("Executing function...")

		// Call the passed function
		functionToRun()

		// Log after function execution
		log.Printf("Function execution completed. Waiting for the next interval.")
	}
}
