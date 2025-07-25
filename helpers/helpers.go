package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file, ignore error if file doesn't exist (optional)
	if err := godotenv.Load(); err != nil {
		log.Warning("⚠️ .env file not found or failed to load")
		log.Error(err.Error())
	}
}

// toSnakeCase converts any input string to UPPER_SNAKE_CASE
func toSnakeCase(input string) string {
	s := strings.ReplaceAll(input, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, "${1}_${2}")

	return strings.ToUpper(s)
}

// GetValue loads the environment variable value for a given key (case insensitive,
// converts input key to UPPER_SNAKE_CASE), including those loaded from .env file.
func GetENVValue(key string) string {
	snakeKey := toSnakeCase(key)
	return os.Getenv(snakeKey)
}

// CreateError returns a new error with the given message
func CreateError(message string) error {
	return errors.New(message)
}

// RespondWithJSON writes a JSON response to the HTTP response writer.
// Constructs a standardized server response with payload and success flag.
// Sets the appropriate headers, status code, and writes the JSON data.
func RespondWithJSON(Response http.ResponseWriter, StatusCode int, Payload interface{}) {
	// Determine success based on whether the status code indicates a client error.
	success := StatusCode < http.StatusBadRequest
	// Log the start of JSON response preparation with status and success details.
	log.Info("📤 Preparing JSON response (status: " + http.StatusText(StatusCode) + ", success: " + boolToStr(success) + ")")

	// Construct the server response with the provided payload and success flag.
	ResponseData := &models.ServerResponse{
		Message:    Payload,
		Success:    success,
		StatusCode: StatusCode,
	}

	// Marshal the response data to JSON.
	ResponseDataJSON, err := json.Marshal(ResponseData)
	if err != nil {
		// Log and set HTTP 500 status if JSON marshaling fails.
		log.Error("❌ Failed to marshal JSON response: " + err.Error())
		Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate JSON response.
	Response.Header().Set("Content-Type", "application/json")
	// Write the HTTP status code to the response.
	Response.WriteHeader(StatusCode)

	// Write the JSON data to the response writer.
	if _, writeErr := Response.Write(ResponseDataJSON); writeErr != nil {
		// Log if writing the response fails.
		log.Error("❌ Failed to write JSON response: " + writeErr.Error())
	} else {
		// Log successful response delivery.
		log.Success("✅ JSON response sent successfully")
	}
}

// boolToStr returns "true" or "false" for boolean values.
// Used for logging boolean values as strings.
func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
