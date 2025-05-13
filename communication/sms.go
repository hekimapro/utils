package communication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	// Initialize an empty SMS response struct to store the API response
	var smsResponse models.SMSResponse

	// Marshal the SMS payload into JSON format for the request body
	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Log the error and return a wrapped error for context
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to marshal payload into JSON: %w", err)
	}

	// Create a new HTTP POST request with the JSON payload
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// Log the error and return a wrapped error for context
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set required headers for the Africa's Talking API request
	req.Header.Set("apiKey", payload.ATAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Initialize an HTTP client to send the request
	client := &http.Client{}
	// Execute the HTTP request and capture the response
	resp, err := client.Do(req)
	if err != nil {
		// Log the error and return a wrapped error for context
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to execute POST request: %w", err)
	}
	// Ensure the response body is closed after processing
	defer resp.Body.Close()

	// Verify the response status code is in the successful range (2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Return an error with the HTTP status for non-successful responses
		return smsResponse, fmt.Errorf("received non-2xx status code: %s", resp.Status)
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Log the error and return a wrapped error for context
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the response body into the SMS response struct
	err = json.Unmarshal(body, &smsResponse)
	if err != nil {
		// Log the error and return a wrapped error for context
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to parse response body into SMSResponse: %w", err)
	}

	// Return the parsed SMS response on success
	return smsResponse, nil
}