package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
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
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// Determine success based on whether the status code indicates a client error.
	success := statusCode < http.StatusBadRequest

	// Pick a default message from the status code
	message := http.StatusText(statusCode)
	if message == "" {
		message = "Unknown status"
	}

	// Log the start of JSON response preparation with status and success details.
	log.Info("📤 Preparing JSON response (status: " + message + ", success: " + boolToStr(success) + ")")

	// Construct the server response with the provided payload and success flag.
	responseData := &models.ServerResponse{
		Data:    payload,
		Success: success,
		Message: message,
	}

	// Marshal the response data to JSON.
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		log.Error("❌ Failed to marshal JSON response: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Write the JSON data to the response writer.
	if _, writeErr := w.Write(responseJSON); writeErr != nil {
		log.Error("❌ Failed to write JSON response: " + writeErr.Error())
	} else {
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

func ConvertToUUID(dataString string) uuid.UUID {
	if dataString == "" {
		return uuid.Nil
	}
	parsedUUID, err := uuid.Parse(dataString)
	if err != nil {
		return uuid.Nil // return empty UUID if parsing fails
	}

	return parsedUUID
}

func GetQueryUUID(request *http.Request) uuid.UUID {
	ID := request.URL.Query().Get("id")
	return ConvertToUUID(ID)
}

func GetSearchKeyword(request *http.Request) string {
	keyword := request.URL.Query().Get("keyword")
	return keyword
}

func GetUUIDContextData(request *http.Request, key models.ContextKey) uuid.UUID {
	contextValue := request.Context().Value(key)
	dataID, ok := contextValue.(string)
	if !ok {
		return uuid.Nil
	}

	return ConvertToUUID(dataID)
}

func GetStringContextData(request *http.Request, key models.ContextKey) string {
	dataID, ok := request.Context().Value(key).(string)
	if !ok {
		return ""
	}
	return dataID
}

func IsValidUUID(providedID string) bool {
	if providedID == "" {
		return false
	}
	if _, err := uuid.Parse(providedID); err != nil {
		return false
	}
	return true
}

// Define sets of known image and video extensions
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".tiff": true,
	".webp": true,
	".heic": true, // High-efficiency image format (common on iPhones)
	".svg":  true, // Scalable vector image
	".ico":  true, // Windows icon format
	".raw":  true, // Generic raw image
	".cr2":  true, // Canon RAW
	".nef":  true, // Nikon RAW
	".psd":  true, // Photoshop format
}

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".avi":  true,
	".mkv":  true,
	".flv":  true,
	".wmv":  true,
	".webm": true,
	".mpeg": true,
	".3gp":  true, // Common on older phones
	".m4v":  true, // iTunes video format
	".ts":   true, // Transport stream
	".vob":  true, // DVD video
	".rm":   true, // RealMedia
	".ogv":  true, // Ogg video
}

// Function to determine if the file is an image or video
func GetFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename)) // Extract and lowercase the extension

	if imageExtensions[ext] {
		return "image"
	} else if videoExtensions[ext] {
		return "video"
	}
	return "unknown"
}

func NormalizePhoneNumber(phoneNumber string, toNormal bool) string {
	phoneNumber = strings.TrimSpace(phoneNumber)

	if toNormal {
		return phoneNumber
	}

	// Remove leading zero and add country code
	if strings.HasPrefix(phoneNumber, "0") && len(phoneNumber) == 10 {
		return "255" + phoneNumber[1:]
	}

	// If it already starts with 255 and is 12 digits, return as is
	if strings.HasPrefix(phoneNumber, "255") && len(phoneNumber) == 12 {
		return phoneNumber
	}

	return phoneNumber // fallback (return as-is)
}

func IsZeroUUID(ID uuid.UUID) bool {
	return ID.String() == "00000000-0000-0000-0000-000000000000"
}
