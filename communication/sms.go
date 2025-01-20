package communication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hekimapro/utils/models"
)

// baseURL for the Africa's Talking API endpoint.
var baseURL = "https://api.africastalking.com/version1/messaging/bulk"

// MessageStatusCodes maps specific status codes to their corresponding messages.
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

// GetStatusMessage retrieves the human-readable message corresponding to an HTTP status code.
// It returns "Unknown Status Code" if the code is not found in the predefined map.
func GetStatusMessage(code int) string {
	if message, exists := MessageStatusCodes[code]; exists {
		return message
	}
	return "Unknown Status Code"
}

// SendSMS sends a bulk SMS request to the Africa's Talking API.
// It marshals the payload into JSON, sends a POST request, and returns the response or an error.
func SendSMS(payload models.SMSPayload) (models.SMSResponse, error) {

	var smsResponse models.SMSResponse

	// Marshal the payload into JSON format for sending in the HTTP request body.
	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Log and return an error if marshalling fails.
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to marshal payload into JSON: %w", err)
	}

	// Create a new HTTP POST request with the JSON payload.
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// Log and return an error if the request creation fails.
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the necessary headers for the request.
	req.Header.Set("apiKey", payload.ATAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Initialize the HTTP client and send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Log and return an error if the HTTP request fails.
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to execute POST request: %w", err)
	}
	defer resp.Body.Close() // Ensure that the response body is closed after reading.

	// Check if the response status code indicates a successful request (2xx range).
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Return an error if the response code is outside the successful range.
		return smsResponse, fmt.Errorf("received non-2xx status code: %s", resp.Status)
	}

	// Read the body of the response to obtain the SMS API response data.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Log and return an error if reading the response body fails.
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the response body into the smsResponse object.
	err = json.Unmarshal(body, &smsResponse)
	if err != nil {
		// Log and return an error if unmarshalling fails.
		fmt.Println(err.Error())
		return smsResponse, fmt.Errorf("failed to parse response body into SMSResponse: %w", err)
	}

	// Return the SMS response if everything succeeds.
	return smsResponse, nil
}
