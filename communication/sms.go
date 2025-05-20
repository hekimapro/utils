package communication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
)

// baseURL defines the Africa's Talking API endpoint for bulk SMS messaging
// Points to the version 1 messaging endpoint
var baseURL = "https://api.africastalking.com/version1/messaging/bulk"

// MessageStatusCodes maps Africa's Talking API status codes to human-readable messages
// Provides descriptions for common success and error states
var MessageStatusCodes = map[int]string{
	100: "Processed",
	101: "Sent",
	102: "Queued",
	401: "RiskHold",
	402: "InvalidSenderId",
	403: "InvalidPhoneNumber",
	404: "UnsupportedNumberType",
	405: "InsufficientBalance",
	406: "UserInBlacklist",
	407: "CouldNotRoute",
	409: "DoNotDisturbRejection",
	500: "InternalServerError",
	501: "GatewayError",
	502: "RejectedByGateway",
}

// GetStatusMessage retrieves the human-readable message for a given status code
// Returns the corresponding message from MessageStatusCodes or a default unknown message
func GetStatusMessage(code int) string {
	// Check if the status code exists in the map and return its message
	if message, exists := MessageStatusCodes[code]; exists {
		return message
	}
	// Return a default message for unrecognized status codes
	return "Unknown Status Code"
}

// SendSMS sends a bulk SMS request to the Africa's Talking API
// Marshals the SMS payload, sends a POST request, and parses the response
// Returns the SMS response or an error if the request fails
func SendSMS(payload models.SMSPayload) (models.SMSResponse, error) {
	var smsResponse models.SMSResponse

	log.Info("Marshalling SMS payload to JSON")
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshal payload: %v", err))
		return smsResponse, fmt.Errorf("failed to marshal payload into JSON: %w", err)
	}

	log.Info("Creating HTTP request")
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create HTTP request: %v", err))
		return smsResponse, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers
	req.Header.Set("apiKey", payload.ATAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	log.Info("Headers set for HTTP request")

	client := &http.Client{}
	log.Info("Sending SMS request to Africa's Talking API...")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to send request: %v", err))
		return smsResponse, fmt.Errorf("failed to execute POST request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Warning(fmt.Sprintf("Received non-2xx response: %s", resp.Status))
		return smsResponse, fmt.Errorf("received non-2xx status code: %s", resp.Status)
	}

	log.Info("Reading response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to read response body: %v", err))
		return smsResponse, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Info("Unmarshalling response into SMSResponse struct")
	err = json.Unmarshal(body, &smsResponse)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to unmarshal response: %v", err))
		return smsResponse, fmt.Errorf("failed to parse response body into SMSResponse: %w", err)
	}

	log.Success("SMS sent and response parsed successfully")
	return smsResponse, nil
}
