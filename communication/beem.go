package communication

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
	"github.com/hekimapro/utils/request"
)

var beemBaseURL = "https://apisms.beem.africa/v1/send"
var beemDeliveryResportURL = "https://dlrapi.beem.africa/public/v1/delivery-reports"

func createAuthHeader(apiKey, secretKey string) string {
	auth := apiKey + ":" + secretKey
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encoded
}

func SendBeemSMS(payload *models.BeemSMSPayload) (*models.BeemSMSResponse, error) {

	var response models.BeemSMSResponse

	requestData := models.BeemSMSRequestBody{
		SourceAddr:   payload.SenderName,
		ScheduleTime: payload.ScheduleTime,
		Encoding:     "0",
		Message:      payload.Message,
		Recipients:   payload.Recipients,
	}

	headers := &request.Headers{
		"Authorization": createAuthHeader(payload.APIKey, payload.SecretKey),
	}

	rawData, err := request.Post(beemBaseURL, requestData, headers)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("failed to deserialize response")
	}

	return &response, nil
}

func GetDeliveryStatus(payload *models.BeemSMSDeliveryStatusPayload) (*models.BeemSMSDeliveryStatusResponse, error) {

	var response models.BeemSMSDeliveryStatusResponse

	headers := &request.Headers{
		"Authorization": createAuthHeader(payload.APIKey, payload.SecretKey),
	}

	URL := fmt.Sprintf("%s?dest_addr=%s&request_id=%s", beemDeliveryResportURL, payload.PhoneNumber, payload.RequestID)

	rawData, err := request.Get(URL, headers)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(rawData, &response); err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("failed to deserialize response")
	}

	return &response, nil
}
