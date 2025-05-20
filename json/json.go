package json

import (
	"encoding/json"
	"net/http"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
)

// RespondWithJSON writes a JSON response to the HTTP response writer
// Constructs a standardized server response with payload and success flag
// Sets the appropriate headers, status code, and writes the JSON data
func RespondWithJSON(Response http.ResponseWriter, StatusCode int, Payload interface{}) {
	ResponseData := &models.ServerResponse{
		Message: Payload,
		Success: StatusCode < http.StatusBadRequest,
	}

	ResponseDataJSON, err := json.Marshal(ResponseData)
	if err != nil {
		log.Error("JSON conversion error: " + err.Error())
		Response.WriteHeader(http.StatusInternalServerError)
		return
	}

	Response.Header().Add("Content-Type", "application/json")
	Response.WriteHeader(StatusCode)

	_, writeErr := Response.Write(ResponseDataJSON)
	if writeErr != nil {
		log.Error("Failed to write JSON response: " + writeErr.Error())
	}
}
