package scheduler

import (
	"fmt"
	"time"

	"github.com/hekimapro/utils/log"
)

// RunFunctionAtInterval schedules a function to run at regular intervals
// Executes the provided function repeatedly after the specified duration
// Supports optional immediate execution before the first interval
func RunFunctionAtInterval(functionToRun func(), interval time.Duration, runInstant bool) {
	log.Info(fmt.Sprintf("Scheduler started: Function will run every %v. Run instantly: %v", interval, runInstant))

	if runInstant {
		log.Info("Executing function immediately before first interval...")
		functionToRun()
		log.Info("Initial execution completed. Waiting for interval...")
	}

	for {

		log.Info(fmt.Sprintf("Waiting for the next interval of %v...", interval))
		time.Sleep(interval)

		log.Info("Executing function...")
		functionToRun()
		log.Info("Function execution completed.")
	}
}
