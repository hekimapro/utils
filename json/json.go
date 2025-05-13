package json

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hekimapro/utils/models"
)

// RespondWithJSON writes a JSON response to the HTTP response writer
// Constructs a standardized server response with payload and success flag
// Sets the appropriate headers, status code, and writes the JSON data
func RespondWithJSON(Response http.ResponseWriter, StatusCode int, Payload interface{}) {
	// Create a ServerResponse object with the provided payload
	// Set success flag based on whether the status code is in the 2xx range
	ResponseData := &models.ServerResponse{
		Message: Payload,                            // Include the payload in the response
		Success: StatusCode < http.StatusBadRequest, // Success is true for 2xx status codes
	}

	// Marshal the response data to JSON format
	ResponseDataJSON, err := json.Marshal(ResponseData)
	if err != nil {
		// Log the JSON marshaling error and return a 500 status code
		log.Printf("JSON conversion error: %v", err.Error())
		Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate JSON response
	Response.Header().Add("Content-Type", "application/json")

	// Set the HTTP status code for the response
	Response.WriteHeader(StatusCode)

	// Write the JSON data to the response writer
	Response.Write(ResponseDataJSON)
}
