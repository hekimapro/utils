package request

import (
	"bytes"         // bytes provides utilities for creating byte buffers.
	"encoding/json" // json provides JSON encoding and decoding functions.

	// errors provides utilities for creating errors.
	"io"       // io provides interfaces for I/O operations.
	"net/http" // http provides utilities for HTTP requests and responses.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// Headers type alias for map[string]string to store HTTP headers.
type Headers map[string]string

// defaultHeaders returns a map of default HTTP headers.
// Includes Accept and Content-Type set to application/json.
func defaultHeaders() Headers {
	// Initialize default headers for JSON communication.
	return Headers{
		"Accept":       "application/json", // Expect JSON responses.
		"Content-Type": "application/json", // Send JSON request bodies.
	}
}

// mergeHeaders combines default headers with user-provided headers.
// User headers override default headers if they conflict.
func mergeHeaders(userHeaders *Headers) Headers {
	// Start with default headers.
	headers := defaultHeaders()
	// Add or override with user-provided headers if provided.
	if userHeaders != nil {
		for headerKey, headerValue := range *userHeaders {
			headers[headerKey] = headerValue
		}
	}
	return headers
}

// Get sends an HTTP GET request to the specified URL.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Get(url string, headers *Headers) (json.RawMessage, error) {
	// Log the start of the GET request.
	log.Info("🔍 Preparing GET request to " + url)
	// Create a new HTTP GET request.
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create GET request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request.
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Log and return an error if the request fails.
		log.Error("❌ GET request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}

// Post sends an HTTP POST request with a JSON body to the specified URL.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Post(url string, body any, headers *Headers) (json.RawMessage, error) {
	// Log the start of the POST request.
	log.Info("📤 Preparing POST request to " + url)

	// Prepare the request body if provided.
	var requestBody io.Reader
	if body != nil {
		// Marshal the body to JSON.
		jsonBody, err := json.Marshal(body)
		if err != nil {
			// Log and return an error if marshaling fails.
			log.Error("❌ Failed to marshal POST body: " + err.Error())
			return nil, err
		}
		// Create a buffer for the JSON body.
		requestBody = bytes.NewBuffer(jsonBody)
	}

	// Create a new HTTP POST request.
	request, err := http.NewRequest(http.MethodPost, url, requestBody)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create POST request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request.
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Log and return an error if the request fails.
		log.Error("❌ POST request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}

// handleResponse processes an HTTP response.
// Reads the body, checks the status code, and returns json.RawMessage or an error.
func handleResponse(response *http.Response) (json.RawMessage, error) {
	// Read the response body.
	body, err := io.ReadAll(response.Body)
	if err != nil {
		// Log and return an error if reading the body fails.
		log.Error("❌ Failed to read response body: " + err.Error())
		return nil, err
	}

	// Unmarshal the body into a json.RawMessage.
	var raw json.RawMessage

	// Check if the status code indicates an error (not 2xx).
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if err := json.Unmarshal(body, &raw); err != nil {
			// Log and return an error if JSON unmarshaling fails.
			log.Error("❌ Failed to unmarshal response JSON: " + err.Error())
			return nil, err
		} else {
			return raw, nil
		}
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		// Log and return an error if JSON unmarshaling fails.
		log.Error("❌ Failed to unmarshal response JSON: " + err.Error())
		return nil, err
	}

	// Log successful response processing.
	log.Info("✅ HTTP response processed successfully")
	return raw, nil
}
