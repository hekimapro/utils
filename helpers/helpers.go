package helpers

import (
	"context"        // context provides support for cancellation and timeouts.
	"crypto/rand"    // rand provides cryptographically secure random number generation.
	"encoding/json"  // json provides JSON encoding and decoding functions.
	"errors"         // errors provides utilities for creating errors.
	"fmt"            // fmt provides formatting and printing functions.
	"io"             // io provides I/O interfaces for file operations.
	"math/big"       // big provides arbitrary-precision arithmetic.
	"mime/multipart" // multipart provides MIME multipart parsing.
	"net/http"       // http provides utilities for HTTP requests and responses.
	"net/url"        // url provides URL parsing and query string manipulation.
	"os"             // os provides file system and environment variable operations.
	"path/filepath"  // filepath provides utilities for file path manipulation.
	"regexp"         // regexp provides regular expression functionality.
	"strconv"        // strconv provides string conversion utilities.
	"strings"        // strings provides string manipulation utilities.
	"time"           // time provides functionality for handling time and durations.

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"            // uuid provides UUID generation and parsing.
	"github.com/hekimapro/utils/log"    // log provides colored logging utilities.
	"github.com/hekimapro/utils/models" // models provides data structures for server responses.
	"github.com/jinzhu/inflection"
	"github.com/joho/godotenv" // godotenv provides .env file loading.
)

func init() {
	// Load .env file, ignore error if file doesn't exist (optional)
	if err := godotenv.Load(); err != nil {
		log.Warning("⚠️ .env file not found or failed to load")
		log.Error(err.Error())
	}
}

// GetENVValue loads the environment variable value for a given key (case insensitive,
// converts input key to UPPER_SNAKE_CASE), including those loaded from .env file.
func GetENVValue(key string) string {
	snakeKey := strings.ToUpper(ToSnakeCase(key))
	return os.Getenv(snakeKey)
}

// GetENVValueWithDefault loads an environment variable with a default value if not set.
func GetENVValueWithDefault(key string, defaultValue string) string {
	if value := GetENVValue(key); value != "" {
		return value
	}
	return defaultValue
}

// GetENVIntValue loads an environment variable as an integer with a default value.
func GetENVIntValue(key string, defaultValue int) int {
	valueStr := GetENVValue(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Warning(fmt.Sprintf("⚠️ Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue))
		return defaultValue
	}

	return value
}

// GetENVBoolValue loads an environment variable as a boolean with a default value.
func GetENVBoolValue(key string, defaultValue bool) bool {
	valueStr := GetENVValue(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Warning(fmt.Sprintf("⚠️ Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue))
		return defaultValue
	}

	return value
}

// CreateError returns a new error with the given message
func CreateError(message string) error {
	return errors.New(message)
}

// CreateErrorf returns a new formatted error with the given message and arguments
func CreateErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// WrapError wraps an existing error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return CreateError(message)
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorf wraps an existing error with formatted additional context
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return CreateErrorf(format, args...)
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// RespondWithJSON writes a JSON response to the HTTP response writer.
// Constructs a standardized server response with payload and success flag.
// Sets the appropriate headers, status code, and writes the JSON data.
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// Determine success based on whether the status code indicates a client error.
	success := statusCode < http.StatusBadRequest

	// // Pick a default message from the status code
	// message := http.StatusText(statusCode)
	// if message == "" {
	// 	message = "Unknown status"
	// }

	// Log the start of JSON response preparation with status and success details.
	log.Info("📤 Preparing JSON response (status: " + http.StatusText(statusCode) + ", success: " + boolToStr(success) + ")")

	// Construct the server response with the provided payload and success flag.
	responseData := &models.ServerResponse{
		Success: success,
		Message: payload,
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

// ConvertToUUID converts a string to UUID, returns uuid.Nil if conversion fails.
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

// GenerateUUID generates a new UUID v4.
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// GenerateUUIDString generates a new UUID v4 as string.
func GenerateUUIDString() string {
	return uuid.New().String()
}

// GetQueryUUID extracts UUID from query parameter "id".
func GetQueryUUID(request *http.Request) uuid.UUID {
	ID := request.URL.Query().Get("id")
	return ConvertToUUID(ID)
}

// GetQueryParam extracts a string query parameter with optional default value.
func GetQueryParam(request *http.Request, key string, defaultValue string) string {
	value := request.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryInt extracts an integer query parameter with optional default value.
func GetQueryInt(request *http.Request, key string, defaultValue int) int {
	value := request.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetQueryBool extracts a boolean query parameter with optional default value.
func GetQueryBool(request *http.Request, key string, defaultValue bool) bool {
	value := request.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// GetSearchKeyword extracts search keyword from query parameter "keyword".
func GetSearchKeyword(request *http.Request) string {
	keyword := request.URL.Query().Get("keyword")
	return keyword
}

// GetUUIDContextData extracts UUID from request context.
func GetUUIDContextData(request *http.Request, key models.ContextKey) uuid.UUID {
	contextValue := request.Context().Value(key)
	dataID, ok := contextValue.(string)
	if !ok {
		return uuid.Nil
	}

	return ConvertToUUID(dataID)
}

// GetStringContextData extracts string from request context.
func GetStringContextData(request *http.Request, key models.ContextKey) string {
	dataID, ok := request.Context().Value(key).(string)
	if !ok {
		return ""
	}
	return dataID
}

// GetIntContextData extracts integer from request context.
func GetIntContextData(request *http.Request, key models.ContextKey) int {
	contextValue := request.Context().Value(key)
	switch v := contextValue.(type) {
	case int:
		return v
	case string:
		if intValue, err := strconv.Atoi(v); err == nil {
			return intValue
		}
	}
	return 0
}

// IsValidUUID checks if the provided string is a valid UUID.
func IsValidUUID(providedID string) bool {
	if providedID == "" {
		return false
	}
	if _, err := uuid.Parse(providedID); err != nil {
		return false
	}
	return true
}

// GetQueryID is an alias for GetQueryUUID for backward compatibility.
func GetQueryID(request *http.Request) uuid.UUID {
	ID := request.URL.Query().Get("id")
	return ConvertToUUID(ID)
}

// GetContextData extracts UUID from request context (alias for GetUUIDContextData).
func GetContextData(request *http.Request, key models.ContextKey) uuid.UUID {
	contextValue := request.Context().Value(key)
	dataID, ok := contextValue.(string)
	if !ok {
		return uuid.Nil
	}

	return ConvertToUUID(dataID)
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

var documentExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".txt":  true,
	".rtf":  true,
	".csv":  true,
}

// GetFileType determines if the file is an image, video, document, or unknown.
func GetFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename)) // Extract and lowercase the extension

	if imageExtensions[ext] {
		return "image"
	} else if videoExtensions[ext] {
		return "video"
	} else if documentExtensions[ext] {
		return "document"
	}
	return "unknown"
}

// IsImageFile checks if the file is an image based on extension.
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExtensions[ext]
}

// IsVideoFile checks if the file is a video based on extension.
func IsVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExtensions[ext]
}

// IsDocumentFile checks if the file is a document based on extension.
func IsDocumentFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return documentExtensions[ext]
}

// NormalizePhoneNumber normalizes phone numbers to standard format.
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

// ValidatePhoneNumber validates phone number format.
func ValidatePhoneNumber(phoneNumber string) bool {
	// Remove any non-digit characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phoneNumber, "")

	// Check if it's a valid length (10 digits for local, 12 for international)
	return len(cleaned) == 10 || len(cleaned) == 12
}

// IsZeroUUID checks if UUID is the zero value.
func IsZeroUUID(ID uuid.UUID) bool {
	return ID.String() == "00000000-0000-0000-0000-000000000000"
}

// GenerateOTP generates a random 6-digit OTP.
func GenerateOTP() (int, error) {
	// The maximum value for a 6-digit number is 999999
	max := big.NewInt(1000000)

	// Generate a random number between 0 and 999999
	n, err := rand.Int(rand.Reader, max)

	if err != nil {
		log.Error(err.Error())
		return 0, CreateError("failed to generate OTP")
	}

	// Ensure the number is 6 digits by adding leading zeros if necessary
	otp := int(n.Int64())
	return otp, nil
}

// GenerateSecureOTPString generates a secure random OTP as string with fixed length.
func GenerateSecureOTPString(length int) (string, error) {
	if length <= 0 {
		return "", CreateError("OTP length must be positive")
	}

	const digits = "0123456789"
	bytes := make([]byte, length)

	for i := range bytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", WrapError(err, "failed to generate secure OTP")
		}
		bytes[i] = digits[num.Int64()]
	}

	return string(bytes), nil
}

// GenerateRandomString generates a cryptographically secure random string of specified length.
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", CreateError("length must be positive")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)

	for i := range bytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", WrapError(err, "failed to generate random string")
		}
		bytes[i] = charset[num.Int64()]
	}

	return string(bytes), nil
}

// ValidateEmail validates an email address format.
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateURL validates a URL format.
func ValidateURL(urlString string) bool {
	_, err := url.ParseRequestURI(urlString)
	return err == nil
}

// GetPaginationParams extracts pagination parameters from request.
func GetPaginationParams(request *http.Request) (page int, pageSize int) {
	page = GetQueryInt(request, "page", 1)
	pageSize = GetQueryInt(request, "pageSize", 10)

	// Ensure minimum values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Limit maximum page size
	}

	return page, pageSize
}

// CalculateOffset calculates database offset for pagination.
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

// GetClientIP extracts client IP address from request.
func GetClientIP(request *http.Request) string {
	// Check for forwarded IP first (behind proxy)
	if ip := request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to remote address
	return strings.Split(request.RemoteAddr, ":")[0]
}

// GetUserAgent extracts user agent from request.
func GetUserAgent(request *http.Request) string {
	return request.Header.Get("User-Agent")
}

// SafeDeleteFile safely deletes a file with error handling.
func SafeDeleteFile(filePath string) error {
	if filePath == "" {
		return CreateError("file path cannot be empty")
	}

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			log.Warning(fmt.Sprintf("⚠️ File not found during deletion: %s", filePath))
			return nil // File doesn't exist, consider it deleted
		}
		return WrapError(err, "failed to delete file")
	}

	log.Info(fmt.Sprintf("🗑️  File deleted successfully: %s", filePath))
	return nil
}

// EnsureDirectory creates directory if it doesn't exist.
func EnsureDirectory(dirPath string) error {
	if dirPath == "" {
		return CreateError("directory path cannot be empty")
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return WrapError(err, "failed to create directory")
	}

	return nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists.
func DirectoryExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetFileSize returns file size in bytes.
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, WrapError(err, "failed to get file size")
	}
	return info.Size(), nil
}

// ReadFileContent reads entire file content as string.
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", WrapError(err, "failed to read file")
	}
	return string(content), nil
}

// WriteFileContent writes string content to file.
func WriteFileContent(filePath string, content string) error {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return WrapError(err, "failed to write file")
	}
	return nil
}

// ParseMultipartForm parses multipart form with size limits.
func ParseMultipartForm(request *http.Request, maxMemory int64) error {
	if err := request.ParseMultipartForm(maxMemory); err != nil {
		return WrapError(err, "failed to parse multipart form")
	}
	return nil
}

// GetFormFile retrieves file from multipart form.
func GetFormFile(request *http.Request, fieldName string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := request.FormFile(fieldName)
	if err != nil {
		return nil, nil, WrapError(err, "failed to get form file")
	}
	return file, header, nil
}

// SaveUploadedFile saves uploaded file to specified path.
func SaveUploadedFile(fileHeader *multipart.FileHeader, destination string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return WrapError(err, "failed to open uploaded file")
	}
	defer file.Close()

	// Create destination file
	destFile, err := os.Create(destination)
	if err != nil {
		return WrapError(err, "failed to create destination file")
	}
	defer destFile.Close()

	// Copy file content
	if _, err := io.Copy(destFile, file); err != nil {
		return WrapError(err, "failed to copy file content")
	}

	return nil
}

// GetCurrentTimestamp returns current timestamp in various formats.
func GetCurrentTimestamp() time.Time {
	return time.Now()
}

// GetCurrentTimestampString returns current timestamp as string in RFC3339 format.
func GetCurrentTimestampString() string {
	return time.Now().Format(time.RFC3339)
}

// FormatTimestamp formats timestamp to specified layout.
func FormatTimestamp(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseTimestamp parses timestamp string with specified layout.
func ParseTimestamp(timestamp string, layout string) (time.Time, error) {
	parsed, err := time.Parse(layout, timestamp)
	if err != nil {
		return time.Time{}, WrapError(err, "failed to parse timestamp")
	}
	return parsed, nil
}

// AddToContext adds a key-value pair to context and returns new context.
func AddToContext(ctx context.Context, key models.ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetFromContext retrieves value from context with type assertion.
func GetFromContext(ctx context.Context, key models.ContextKey) (interface{}, bool) {
	value := ctx.Value(key)
	return value, value != nil
}

// GetStringFromContext retrieves string value from context.
func GetStringFromContext(ctx context.Context, key models.ContextKey) string {
	value, ok := GetFromContext(ctx, key)
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

// GetIntFromContext retrieves integer value from context.
func GetIntFromContext(ctx context.Context, key models.ContextKey) int {
	value, ok := GetFromContext(ctx, key)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case string:
		if intValue, err := strconv.Atoi(v); err == nil {
			return intValue
		}
	}
	return 0
}

// ContainsString checks if slice contains string.
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsInt checks if slice contains integer.
func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate values from string slice.
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Coalesce returns the first non-empty string from the provided arguments.
func Coalesce(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// DefaultIfEmpty returns the value if not empty, otherwise returns default value.
func DefaultIfEmpty(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// TruncateString truncates string to specified length with ellipsis.
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength < 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}

func ToFormatedCurrency(value float64) string {
	return humanize.CommafWithDigits(value, 0)
}

func ToPlural(word string) string {
	return inflection.Plural(word)
}

func ToSingular(word string) string {
	return inflection.Singular(word)
}

func ToSnakeCase(input string) string {
	input = strings.TrimSpace(input)

	// Replace spaces and hyphens with underscore
	input = regexp.MustCompile(`[\s\-]+`).ReplaceAllString(input, "_")

	// Insert underscore before capital letters
	input = regexp.MustCompile(`([a-z0-9])([A-Z])`).ReplaceAllString(input, "${1}_${2}")

	// Convert to lower case
	return strings.ToLower(input)
}


