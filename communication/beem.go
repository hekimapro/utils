package communication

import (
	"encoding/base64" // base64 provides functions for encoding authentication credentials.
	"encoding/json"   // json provides functions for JSON encoding and decoding.
	"fmt"             // fmt provides formatting and printing functions.

	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
	"github.com/hekimapro/utils/models"  // models contains data structures for API payloads and responses.
	"github.com/hekimapro/utils/request" // request provides utilities for making HTTP requests.
)

// beemBaseURL defines the Beem API endpoint for sending SMS messages.
var beemBaseURL = "https://apisms.beem.africa/v1/send"

// beemDeliveryResportURL defines the Beem API endpoint for retrieving SMS delivery reports.
// Note: "Resport" is likely a typo in the original code and should be "Report".
var beemDeliveryResportURL = "https://dlrapi.beem.africa/public/v1/delivery-reports"

// createAuthHeader generates a Base64-encoded Authorization header for Beem API requests.
// Combines API key and secret key into a Basic Authentication string.
func createAuthHeader(apiKey, secretKey string) string {
	// Concatenate API key and secret key with a colon separator.
	auth := apiKey + ":" + secretKey
	// Encode the concatenated string to Base64.
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	// Prefix with "Basic " for HTTP Basic Authentication.
	return "Basic " + encoded
}

// SendBeemSMS sends an SMS request to the Beem API.
// Constructs the request payload, sends a POST request, and parses the response.
// Returns the SMS response or an error if the request fails.
func SendBeemSMS(payload *models.BeemSMSPayload) (*models.BeemSMSResponse, error) {
	var response models.BeemSMSResponse

	// Construct the request body with payload details.
	requestData := models.BeemSMSRequestBody{
		SourceAddr:   payload.SenderName,   // Sender name for the SMS.
		ScheduleTime: payload.ScheduleTime, // Optional scheduling time for the SMS.
		Encoding:     "0",                 // Default encoding (plain text).
		Message:      payload.Message,     // SMS message content.
		Recipients:   payload.Recipients,  // List of recipient phone numbers.
	}

	// Set Authorization header using API key and secret key.
	headers := &request.Headers{
		"Authorization": createAuthHeader(payload.APIKey, payload.SecretKey),
	}

	// Send POST request to Beem API with the constructed payload and headers.
	rawData, err := request.Post(beemBaseURL, requestData, headers)
	if err != nil {
		log.Error(err.Error()) // Log error if the request fails.
		return nil, err
	}

	// Deserialize the raw response into the BeemSMSResponse struct.
	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error()) // Log error if deserialization fails.
		return nil, fmt.Errorf("failed to deserialize response")
	}

	return &response, nil
}

// GetDeliveryStatus retrieves the delivery status of an SMS from the Beem API.
// Sends a GET request with query parameters and parses the response.
// Returns the delivery status response or an error if the request fails.
func GetDeliveryStatus(payload *models.BeemSMSDeliveryStatusPayload) (*models.BeemSMSDeliveryStatusResponse, error) {
	var response models.BeemSMSDeliveryStatusResponse

	// Set Authorization header using API key and secret key.
	headers := &request.Headers{
		"Authorization": createAuthHeader(payload.APIKey, payload.SecretKey),
	}

	// Construct the URL with query parameters for phone number and request ID.
	URL := fmt.Sprintf("%s?dest_addr=%s&request_id=%d", beemDeliveryResportURL, payload.PhoneNumber, payload.RequestID)

	// Send GET request to Beem API to fetch delivery status.
	rawData, err := request.Get(URL, headers)
	if err != nil {
		log.Error(err.Error()) // Log error if the request fails.
		return nil, err
	}

	// Deserialize the raw response into the BeemSMSDeliveryStatusResponse struct.
	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error()) // Log error if deserialization fails.
		return nil, fmt.Errorf("failed to deserialize response")
	}

	return &response, nil
}