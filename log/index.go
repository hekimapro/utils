// Package log provides colored logging functions with timestamps, log levels, and structured logging support.
package log

import (
	"context" // context provides support for contextual logging.
	"fmt"     // fmt provides formatting and printing functions.
	"io"      // io provides I/O interfaces for output redirection.
	"os"      // os provides access to standard output and environment variables.
	"runtime" // runtime provides access to stack trace information.
	"strings" // strings provides string manipulation utilities.
	"sync"    // sync provides synchronization primitives for thread safety.
	"time"    // time provides functionality for handling time and timestamps.
)

// Constants for ANSI color codes used in log output formatting.
const (
	reset        = "\033[0m"  // reset resets the terminal color to default.
	brightRed    = "\033[91m" // brightRed is the ANSI code for bright red text.
	brightGreen  = "\033[92m" // brightGreen is the ANSI code for bright green text.
	brightYellow = "\033[93m" // brightYellow is the ANSI code for bright yellow text.
	brightBlue   = "\033[94m" // brightBlue is the ANSI code for bright blue text.
	brightCyan   = "\033[96m" // brightCyan is the ANSI code for bright cyan text.
	brightWhite  = "\033[97m" // brightWhite is the ANSI code for bright white text.
)

// LogLevel represents the severity level of log messages.
type LogLevel int

const (
	LevelDebug   LogLevel = iota // LevelDebug represents debug-level messages
	LevelInfo                    // LevelInfo represents informational messages
	LevelSuccess                 // LevelSuccess represents success messages
	LevelWarning                 // LevelWarning represents warning messages
	LevelError                   // LevelError represents error messages
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelSuccess:
		return "SUCCESS"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig holds configuration for the logger.
type LoggerConfig struct {
	MinLevel     LogLevel  // MinLevel specifies the minimum log level to output
	EnableColors bool      // EnableColors specifies whether to use colored output
	Output       io.Writer // Output specifies the output writer for logs
	EnableCaller bool      // EnableCaller specifies whether to include caller information
	TimeFormat   string    // TimeFormat specifies the timestamp format
}

// globalConfig holds the global logger configuration.
var globalConfig = LoggerConfig{
	MinLevel:     LevelInfo,                   // Default to INFO level and above
	EnableColors: true,                        // Enable colors by default
	Output:       os.Stdout,                   // Output to stdout by default
	EnableCaller: false,                       // Disable caller info by default
	TimeFormat:   "Mon Jan 2006 15:04:05.000", // Default time format
}

var configMutex sync.RWMutex // Mutex for thread-safe configuration changes

// SetConfig updates the global logger configuration.
func SetConfig(config LoggerConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if config.MinLevel >= LevelDebug && config.MinLevel <= LevelError {
		globalConfig.MinLevel = config.MinLevel
	}
	if config.Output != nil {
		globalConfig.Output = config.Output
	}
	globalConfig.EnableColors = config.EnableColors
	globalConfig.EnableCaller = config.EnableCaller
	if config.TimeFormat != "" {
		globalConfig.TimeFormat = config.TimeFormat
	}
}

// SetMinLevel sets the minimum log level for output.
func SetMinLevel(level LogLevel) {
	configMutex.Lock()
	defer configMutex.Unlock()
	if level >= LevelDebug && level <= LevelError {
		globalConfig.MinLevel = level
	}
}

// SetOutput sets the output writer for logs.
func SetOutput(writer io.Writer) {
	configMutex.Lock()
	defer configMutex.Unlock()
	if writer != nil {
		globalConfig.Output = writer
	}
}

// DisableColors disables colored output.
func DisableColors() {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig.EnableColors = false
}

// EnableCallerInfo enables including caller information in logs.
func EnableCallerInfo() {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig.EnableCaller = true
}

// getCallerInfo returns the caller file and line number for logging.
func getCallerInfo() string {
	// Skip 3 callers: getCallerInfo -> logInternal -> public log function
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}

	// Shorten file path to just the file name
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// shouldLog checks if the given log level should be logged based on configuration.
func shouldLog(level LogLevel) bool {
	configMutex.RLock()
	minLevel := globalConfig.MinLevel
	configMutex.RUnlock()

	return level >= minLevel
}

// getColor returns the ANSI color code for the given log level.
func getColor(level LogLevel) string {
	configMutex.RLock()
	enableColors := globalConfig.EnableColors
	configMutex.RUnlock()

	if !enableColors {
		return ""
	}

	switch level {
	case LevelDebug:
		return brightCyan
	case LevelInfo:
		return brightBlue
	case LevelSuccess:
		return brightGreen
	case LevelWarning:
		return brightYellow
	case LevelError:
		return brightRed
	default:
		return brightWhite
	}
}

// extractContextFields extracts relevant fields from context for logging.
// This provides a hook for context-aware logging without changing the function signature.
func extractContextFields(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// This is an extensible hook - you can add more context value extraction here
	// For example, you might extract request ID, user ID, correlation ID, etc.

	// Example implementation that would work with common context patterns:
	// if requestID, ok := ctx.Value("request_id").(string); ok {
	//     return fmt.Sprintf(" [request_id:%s]", requestID)
	// }

	return ""
}

// logInternal is the internal logging function that handles all log output with context support.
func logInternal(level LogLevel, message string) {
	if !shouldLog(level) {
		return
	}

	// Get configuration values
	configMutex.RLock()
	output := globalConfig.Output
	enableColors := globalConfig.EnableColors
	enableCaller := globalConfig.EnableCaller
	timeFormat := globalConfig.TimeFormat
	configMutex.RUnlock()

	// Prepare log components
	timestamp := time.Now().Format(timeFormat)
	levelStr := level.String()
	color := getColor(level)

	// Build additional information string
	var extraInfo strings.Builder

	// Add caller information if enabled
	if enableCaller {
		if callerInfo := getCallerInfo(); callerInfo != "" {
			extraInfo.WriteString(" [")
			extraInfo.WriteString(callerInfo)
			extraInfo.WriteString("]")
		}
	}

	// Add context information (using background context for now)
	// This provides the infrastructure for context-aware logging
	if ctxFields := extractContextFields(context.Background()); ctxFields != "" {
		extraInfo.WriteString(ctxFields)
	}

	// Format the log message
	var logLine string
	if enableColors {
		logLine = fmt.Sprintf("%s[%s] %s %s%s%s\n",
			color, levelStr, timestamp, message, extraInfo.String(), reset)
	} else {
		logLine = fmt.Sprintf("[%s] %s %s%s\n",
			levelStr, timestamp, message, extraInfo.String())
	}

	// Write to output
	fmt.Fprint(output, logLine)
}

// Info logs an informational message with a blue [INFO] prefix and timestamp.
// Now includes internal context support infrastructure.
func Info(message string) {
	logInternal(LevelInfo, message)
}

// Infof logs a formatted informational message.
func Infof(format string, args ...interface{}) {
	logInternal(LevelInfo, fmt.Sprintf(format, args...))
}

// Success logs a success message with a green [SUCCESS] prefix and timestamp.
// Now includes internal context support infrastructure.
func Success(message string) {
	logInternal(LevelSuccess, message)
}

// Successf logs a formatted success message.
func Successf(format string, args ...interface{}) {
	logInternal(LevelSuccess, fmt.Sprintf(format, args...))
}

// Warning logs a warning message with a yellow [WARNING] prefix and timestamp.
// Now includes internal context support infrastructure.
func Warning(message string) {
	logInternal(LevelWarning, message)
}

// Warningf logs a formatted warning message.
func Warningf(format string, args ...interface{}) {
	logInternal(LevelWarning, fmt.Sprintf(format, args...))
}

// Error logs an error message with a red [ERROR] prefix and timestamp.
// Now includes internal context support infrastructure.
func Error(message string) {
	logInternal(LevelError, message)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	logInternal(LevelError, fmt.Sprintf(format, args...))
}

// Debug logs a debug message with a cyan [DEBUG] prefix and timestamp.
// Debug messages are only shown when log level is set to LevelDebug.
// Now includes internal context support infrastructure.
func Debug(message string) {
	logInternal(LevelDebug, message)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	logInternal(LevelDebug, fmt.Sprintf(format, args...))
}

// WithFields creates a structured log entry with additional fields.
// This provides a foundation for structured logging while maintaining simplicity.
func WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		fields: fields,
	}
}

// FieldLogger provides structured logging with additional fields.
type FieldLogger struct {
	fields map[string]interface{}
}

// Info logs an info message with structured fields.
func (f *FieldLogger) Info(message string) {
	f.logWithFields(LevelInfo, message)
}

// Error logs an error message with structured fields.
func (f *FieldLogger) Error(message string) {
	f.logWithFields(LevelError, message)
}

// Warning logs a warning message with structured fields.
func (f *FieldLogger) Warning(message string) {
	f.logWithFields(LevelWarning, message)
}

// Success logs a success message with structured fields.
func (f *FieldLogger) Success(message string) {
	f.logWithFields(LevelSuccess, message)
}

// Debug logs a debug message with structured fields.
func (f *FieldLogger) Debug(message string) {
	f.logWithFields(LevelDebug, message)
}

// logWithFields handles the actual logging with structured fields.
func (f *FieldLogger) logWithFields(level LogLevel, message string) {
	if !shouldLog(level) {
		return
	}

	// Build fields string
	fieldsStr := ""
	if len(f.fields) > 0 {
		for key, value := range f.fields {
			fieldsStr += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	// Log the message with fields
	logInternal(level, message+fieldsStr)
}
