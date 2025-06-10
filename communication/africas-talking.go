package communication

import (
	"encoding/json"
	"fmt"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
	"github.com/hekimapro/utils/request"
)

// baseURL defines the Africa's Talking API endpoint for bulk SMS messaging
// Points to the version 1 messaging endpoint
var ATBaseURL = "https://api.africastalking.com/version1/messaging/bulk"

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
func SendAfricasTalkingSMS(payload *models.ATSMSPayload) (*models.ATSMSResponse, error) {

	var response models.ATSMSResponse

	headers := &request.Headers{
		"apiKey": payload.ATAPIKey,
	}

	rawData, err := request.Post(ATBaseURL, payload, headers)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("failed to deserialize response")
	}
	return &response, nil
}
