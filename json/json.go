package json

import (
	"encoding/json" // json provides functions for JSON encoding.
	"net/http"      // http provides HTTP server and response handling.

	"github.com/hekimapro/utils/log"    // log provides colored logging utilities.
	"github.com/hekimapro/utils/models" // models contains data structures for server responses.
)

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
		Message: Payload,
		Success: success,
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