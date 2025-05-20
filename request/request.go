package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hekimapro/utils/log"
)

// Headers type alias for map[string]string
type Headers map[string]string

// defaultHeaders returns a map of default HTTP headers
// Includes Accept and Content-Type set to application/json
func defaultHeaders() Headers {
	return Headers{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
}

// mergeHeaders combines default headers with user-provided headers
// User headers override default headers if they conflict
func mergeHeaders(userHeaders *Headers) Headers {
	headers := defaultHeaders()
	if userHeaders != nil {
		for headerKey, headerValue := range *userHeaders {
			headers[headerKey] = headerValue
		}
	}
	return headers
}

// Get sends an HTTP GET request to the specified URL
// Applies headers and returns the response body as json.RawMessage
func Get(url string, headers *Headers) (json.RawMessage, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error("Error creating GET request: " + err.Error())
		return nil, err
	}

	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error("Error sending GET request: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	return handleResponse(response)
}

// Post sends an HTTP POST request with a JSON body to the specified URL
// Applies headers and returns the response body as json.RawMessage
func Post(url string, body any, headers *Headers) (json.RawMessage, error) {
	var requestBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Error("Error marshalling POST body: " + err.Error())
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonBody)
	}

	request, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		log.Error("Error creating POST request: " + err.Error())
		return nil, err
	}

	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error("Error sending POST request: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	return handleResponse(response)
}

// handleResponse processes an HTTP response
// Reads the body, checks the status code, and returns json.RawMessage or error
func handleResponse(response *http.Response) (json.RawMessage, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading response body: " + err.Error())
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		errMessage := string(body)
		log.Error("Error in response status: " + http.StatusText(response.StatusCode) + " - " + errMessage)
		return nil, errors.New(errMessage)
	}

	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Error("Error unmarshalling response body: " + err.Error())
		return nil, err
	}

	return raw, nil
}
