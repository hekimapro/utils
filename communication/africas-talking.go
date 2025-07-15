// Package communication handles interactions with external APIs for messaging, such as Africa's Talking SMS service.
package communication

import (
	"encoding/json" // json provides functions for JSON encoding and decoding.
	"fmt"           // fmt provides formatting and printing functions.

	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
	"github.com/hekimapro/utils/models"  // models contains data structures for API payloads and responses.
	"github.com/hekimapro/utils/request" // request provides utilities for making HTTP requests.
)

// baseURL defines the Africa's Talking API endpoint for bulk SMS messaging.
// Points to the version 1 messaging endpoint.
var ATBaseURL = "https://api.africastalking.com/version1/messaging/bulk"

// MessageStatusCodes maps Africa's Talking API status codes to human-readable messages.
// Provides descriptions for common success and error states.
var MessageStatusCodes = map[int]string{
	100: "Processed",            // Message has been processed by the API.
	101: "Sent",                 // Message successfully sent to the recipient.
	102: "Queued",               // Message queued for sending.
	401: "RiskHold",             // Message held due to risk checks.
	402: "InvalidSenderId",      // Invalid sender ID provided.
	403: "InvalidPhoneNumber",   // Invalid recipient phone number.
	404: "UnsupportedNumberType", // Number type not supported by the API.
	405: "InsufficientBalance",   // Insufficient account balance to send message.
	406: "UserInBlacklist",      // Recipient is blacklisted.
	407: "CouldNotRoute",        // Unable to route the message.
	409: "DoNotDisturbRejection", // Message rejected due to Do Not Disturb settings.
	500: "InternalServerError",   // API server encountered an internal error.
	501: "GatewayError",          // Error occurred at the gateway.
	502: "RejectedByGateway",     // Message rejected by the gateway.
}

// GetStatusMessage retrieves the human-readable message for a given status code.
// Returns the corresponding message from MessageStatusCodes or a default unknown message.
func GetStatusMessage(code int) string {
	// Check if the status code exists in the map and return its message.
	if message, exists := MessageStatusCodes[code]; exists {
		return message
	}
	// Return a default message for unrecognized status codes.
	return "Unknown Status Code"
}

// SendAfricasTalkingSMS sends a bulk SMS request to the Africa's Talking API.
// Marshals the SMS payload, sends a POST request, and parses the response.
// Returns the SMS response or an error if the request fails.
func SendAfricasTalkingSMS(payload *models.ATSMSPayload) (*models.ATSMSResponse, error) {
	var response models.ATSMSResponse

	// Set API key in request headers for authentication.
	headers := &request.Headers{
		"apiKey": payload.ATAPIKey,
	}

	// Send POST request to Africa's Talking API with payload and headers.
	rawData, err := request.Post(ATBaseURL, payload, headers)
	if err != nil {
		return nil, err
	}

	// Deserialize the raw response into the ATSMSResponse struct.
	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error()) // Log error if deserialization fails.
		return nil, fmt.Errorf("failed to deserialize response")
	}
	return &response, nil
}