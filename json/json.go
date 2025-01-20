package json

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hekimapro/utils/models"
)

// RespondWithJSON writes a JSON response to the http.ResponseWriter.
// Response: The HTTP response writer to send the response.
// StatusCode: The HTTP status code to indicate the outcome of the request.
// Payload: The data to be included in the response body.
func RespondWithJSON(Response http.ResponseWriter, StatusCode int, Payload interface{}) {
	// Construct the ServerResponse object.
	// Success is determined based on whether the status code is a successful 2xx status code.
	ResponseData := &models.ServerResponse{
		Message: Payload,                            // Set the message or payload of the response.
		Success: StatusCode < http.StatusBadRequest, // Set success flag based on status code (2xx success).
	}

	// Convert the response data to JSON format.
	ResponseDataJSON, err := json.Marshal(ResponseData)
	if err != nil {
		// Log the error and set the HTTP status code to 500 if JSON conversion fails.
		log.Printf("JSON conversion error: %v", err.Error())
		Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate that the response body is in JSON format.
	Response.Header().Add("Content-Type", "application/json")

	// Set the HTTP status code for the response.
	Response.WriteHeader(StatusCode)

	// Write the JSON response to the http.ResponseWriter.
	Response.Write(ResponseDataJSON)
}
