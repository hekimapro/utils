package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hekimapro/utils/models"
)

var ATBaseURL = "https://api.africastalking.com/version1/messaging/bulk"

var StatusCodeMessages = map[int]string{
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

func GetStatusMessage(code int) string {
	if message, exists := StatusCodeMessages[code]; exists {
		return message
	}
	return "Unknown Status Code"
}

func SendSMS(payload models.SMSPayload) (models.SMSResponse, error) {

	var smsResponse models.SMSResponse

	// Marshal the payload into JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err.Error())
		return smsResponse, CreateError("failed to marshal payload")
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", ATBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err.Error())
		return smsResponse, CreateError("failed to create request")
	}

	// Set headers
	req.Header.Set("apiKey", payload.ATAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return smsResponse, CreateError("failed to make post request")
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return smsResponse, CreateError("received non-2xx status code: " + resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return smsResponse, CreateError("failed to read response body")
	}

	// Unmarshal the response body into SMSResponse
	err = json.Unmarshal(body, &smsResponse)
	if err != nil {
		fmt.Println(err.Error())
		return smsResponse, CreateError("failed to parse response body")
	}

	return smsResponse, nil
}
