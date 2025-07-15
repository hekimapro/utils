// Package log provides colored logging functions with timestamps.
package log

import (
	"fmt"  // fmt provides formatting and printing functions.
	"time" // time provides functionality for handling time and timestamps.
)

// Constants for ANSI color codes used in log output formatting.
const (
	reset        = "\033[0m"  // reset resets the terminal color to default.
	brightRed    = "\033[91m" // brightRed is the ANSI code for bright red text.
	brightGreen  = "\033[92m" // brightGreen is the ANSI code for bright green text.
	brightYellow = "\033[93m" // brightYellow is the ANSI code for bright yellow text.
	brightBlue   = "\033[94m" // brightBlue is the ANSI code for bright blue text.
)

// formattedTime returns the current time formatted as a string in the format "Mon Jan 2006 15:04:05.000".
func formattedTime() string {
	return time.Now().Format("Mon Jan 2006 15:04:05.000") // Uses time.Now() to get current time and formats it.
}

// Info logs an informational message with a blue [INFO] prefix and timestamp.
func Info(message string) {
	// Combines blue color, INFO tag, timestamp, message, and color reset in output.
	fmt.Printf("%s[INFO] %s %s%s\n", brightBlue, formattedTime(), message, reset)
}

// Success logs a success message with a green [SUCCESS] prefix and timestamp.
func Success(message string) {
	// Combines green color, SUCCESS tag, timestamp, message, and color reset in output.
	fmt.Printf("%s[SUCCESS] %s %s%s\n", brightGreen, formattedTime(), message, reset)
}

// Warning logs a warning message with a yellow [WARNING] prefix and timestamp.
func Warning(message string) {
	// Combines yellow color, WARNING tag, timestamp, message, and color reset in output.
	fmt.Printf("%s[WARNING] %s %s%s\n", brightYellow, formattedTime(), message, reset)
}

// Error logs an error message with a red [ERROR] prefix and timestamp.
func Error(message string) {
	// Combines red color, ERROR tag, timestamp, message, and color reset in output.
	fmt.Printf("%s[ERROR] %s %s%s\n", brightRed, formattedTime(), message, reset)
}