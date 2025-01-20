package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// ServerResponse represents the structure of the JSON response.
type ServerResponse struct {
	Success bool        `json:"success"` // Indicates whether the operation was successful.
	Message interface{} `json:"message"` // Contains the response message or payload.
}

// RespondWithJSON writes a JSON response to the http.ResponseWriter.
func RespondWithJSON(Response http.ResponseWriter, StatusCode int, Payload interface{}) {
	// Construct the ServerResponse object with the JSON data and success status.
	ResponseData := &ServerResponse{
		Message: Payload,                            // Set the message or payload of the response.
		Success: StatusCode < http.StatusBadRequest, // Determines success based on status code.
	}

	// Convert the response data to JSON format.
	ResponseDataJSON, Error := json.Marshal(ResponseData)
	if Error != nil {
		// Log and set HTTP status code to 500 if JSON conversion fails.
		log.Printf("JSON conversion error: %v", Error.Error())
		Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate JSON data.
	Response.Header().Add("Content-Type", "application/json")

	// Set the HTTP status code for the response.
	Response.WriteHeader(StatusCode)

	// Write the JSON response to the http.ResponseWriter.
	Response.Write(ResponseDataJSON)
}
