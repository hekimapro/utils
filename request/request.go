package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
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
	// Start with default headers
	headers := defaultHeaders()
	// Add or override with user-provided headers if provided
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
	// Create a new GET request
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// Log and return error if request creation fails
		log.Printf("Error creating GET request: %v", err)
		return nil, err
	}

	// Apply merged headers to the request
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Send the request using the default HTTP client
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Log and return error if request execution fails
		log.Printf("Error sending GET request: %v", err)
		return nil, err
	}
	// Ensure response body is closed after processing
	defer response.Body.Close()

	// Handle the response and return the JSON body or error
	return handleResponse(response)
}

// Post sends an HTTP POST request with a JSON body to the specified URL
// Applies headers and returns the response body as json.RawMessage
func Post(url string, body any, headers *Headers) (json.RawMessage, error) {
	// Initialize request body as nil
	var requestBody io.Reader

	// Marshal the body to JSON if provided
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			// Log and return error if JSON marshaling fails
			log.Printf("Error marshalling POST body: %v", err)
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonBody)
	}

	// Create a new POST request with the JSON body
	request, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		// Log and return error if request creation fails
		log.Printf("Error creating POST request: %v", err)
		return nil, err
	}

	// Apply merged headers to the request
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Send the request using the default HTTP client
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Log and return error if request execution fails
		log.Printf("Error sending POST request: %v", err)
		return nil, err
	}
	// Ensure response body is closed after processing
	defer response.Body.Close()

	// Handle the response and return the JSON body or error
	return handleResponse(response)
}

// handleResponse processes an HTTP response
// Reads the body, checks the status code, and returns json.RawMessage or error
func handleResponse(response *http.Response) (json.RawMessage, error) {
	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		// Log and return error if reading the body fails
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	// Check for non-2xx status codes
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// Log and return error with status code and body message
		errMessage := string(body)
		log.Printf("Error in response status: %v - %v", response.StatusCode, errMessage)
		return nil, errors.New(errMessage)
	}

	// Unmarshal the body to json.RawMessage
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		// Log and return error if unmarshaling fails
		log.Printf("Error unmarshalling response body: %v", err)
		return nil, err
	}

	// Return the raw JSON response
	return raw, nil
}
