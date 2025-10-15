package request

import (
	"bytes"         // bytes provides utilities for creating byte buffers.
	"context"       // context provides support for cancellation and timeouts.
	"encoding/json" // json provides JSON encoding and decoding functions.
	"errors"        // errors provides utilities for creating errors.
	"fmt"           // fmt provides formatting and printing functions.
	"io"            // io provides interfaces for I/O operations.
	"net/http"      // http provides utilities for HTTP requests and responses.
	"time"          // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// Headers type alias for map[string]string to store HTTP headers.
type Headers map[string]string

// RequestConfig holds configuration parameters for HTTP requests.
// This struct centralizes all request settings for better maintainability.
type RequestConfig struct {
	Timeout    time.Duration // Timeout specifies the maximum time for the entire request
	MaxRetries int           // MaxRetries specifies maximum retry attempts for failed requests
	RetryDelay time.Duration // RetryDelay specifies the delay between retry attempts
}

// LoadConfig loads request configuration with defaults.
// Returns a RequestConfig struct with default values.
func LoadConfig() RequestConfig {
	return RequestConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// validateURL validates that the URL is not empty.
// Returns an error if the URL is empty.
func validateURL(url string) error {
	if url == "" {
		return errors.New("URL cannot be empty")
	}
	return nil
}

// createHTTPClient creates an HTTP client with configured timeouts.
// Returns an *http.Client with the specified timeout.
func createHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		// Add transport with additional settings if needed
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

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

// shouldRetry determines if a request should be retried based on status code and error.
// Returns true for network errors and 5xx status codes.
func shouldRetry(statusCode int, err error) bool {
	if err != nil {
		return true // Retry on network errors
	}

	// Retry on server errors (5xx) and 429 (Too Many Requests)
	if statusCode >= 500 || statusCode == http.StatusTooManyRequests {
		return true
	}

	return false
}

// executeWithRetry executes an HTTP request with retry logic and context support.
// Returns the HTTP response or an error after all retry attempts.
func executeWithRetry(ctx context.Context, req *http.Request, config RequestConfig) (*http.Response, error) {
	var lastError error
	var response *http.Response

	client := createHTTPClient(config.Timeout)

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check if context is cancelled before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue with request
		}

		// Log retry attempt if not the first attempt
		if attempt > 0 {
			log.Warning(fmt.Sprintf("🔄 Retry attempt %d/%d for %s %s",
				attempt, config.MaxRetries, req.Method, req.URL.String()))
			time.Sleep(config.RetryDelay * time.Duration(attempt)) // Exponential backoff
		}

		// Execute the request with context
		reqWithContext := req.WithContext(ctx)
		resp, err := client.Do(reqWithContext)

		if err != nil {
			lastError = err
			log.Warning(fmt.Sprintf("⚠️  Request attempt %d failed: %v", attempt+1, err))
			continue
		}

		// Check if we should retry based on status code
		if shouldRetry(resp.StatusCode, nil) {
			lastError = fmt.Errorf("server returned %d status", resp.StatusCode)
			resp.Body.Close()
			log.Warning(fmt.Sprintf("⚠️  Request attempt %d failed with status: %d", attempt+1, resp.StatusCode))
			continue
		}

		// Success - return the response
		return resp, nil
	}

	return response, fmt.Errorf("request failed after %d attempts: %w", config.MaxRetries+1, lastError)
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

	// Log response status and size
	log.Info(fmt.Sprintf("📥 Response received - Status: %d, Size: %d bytes",
		response.StatusCode, len(body)))

	// Check if the status code indicates an error (not 2xx).
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Warning(fmt.Sprintf("⚠️  HTTP error response: %d %s",
			response.StatusCode, http.StatusText(response.StatusCode)))

		// Try to parse error response as JSON
		var errorResponse json.RawMessage
		if json.Unmarshal(body, &errorResponse) == nil {
			return errorResponse, fmt.Errorf("HTTP %d: %s", response.StatusCode, http.StatusText(response.StatusCode))
		}

		// Return raw body if not JSON
		return body, fmt.Errorf("HTTP %d: %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	// Try to unmarshal the body into json.RawMessage for successful responses
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		// Log and return an error if JSON unmarshaling fails.
		log.Error("❌ Failed to unmarshal response JSON: " + err.Error())
		return nil, err
	}

	// Log successful response processing.
	log.Success("✅ HTTP response processed successfully")
	return raw, nil
}

// Get sends an HTTP GET request to the specified URL with context support.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Get(url string, headers *Headers) (json.RawMessage, error) {
	return getWithContext(context.Background(), url, headers)
}

// getWithContext sends an HTTP GET request with context support for cancellation.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func getWithContext(ctx context.Context, url string, headers *Headers) (json.RawMessage, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		log.Error("❌ Invalid URL: " + err.Error())
		return nil, err
	}

	// Load configuration
	config := LoadConfig()

	// Log the start of the GET request with configuration details.
	log.Info(fmt.Sprintf("🔍 Preparing GET request to %s (Timeout: %v, MaxRetries: %d)",
		url, config.Timeout, config.MaxRetries))

	// Create a new HTTP GET request with background context (will be overridden in executeWithRetry)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create GET request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request with retry logic and context.
	response, err := executeWithRetry(ctx, request, config)
	if err != nil {
		// Check if error is due to context cancellation
		if errors.Is(err, context.Canceled) {
			log.Warning("⚠️  GET request canceled by context")
			return nil, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("⏰ GET request timed out")
			return nil, err
		}
		// Log and return an error if the request fails.
		log.Error("❌ GET request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}

// Post sends an HTTP POST request with a JSON body to the specified URL with context support.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Post(url string, body any, headers *Headers) (json.RawMessage, error) {
	return postWithContext(context.Background(), url, body, headers)
}

// postWithContext sends an HTTP POST request with context support for cancellation.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func postWithContext(ctx context.Context, url string, body any, headers *Headers) (json.RawMessage, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		log.Error("❌ Invalid URL: " + err.Error())
		return nil, err
	}

	// Load configuration
	config := LoadConfig()

	// Log the start of the POST request.
	log.Info(fmt.Sprintf("📤 Preparing POST request to %s (Timeout: %v, MaxRetries: %d)",
		url, config.Timeout, config.MaxRetries))

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

		// Log request body size for debugging
		log.Info(fmt.Sprintf("📦 Request body size: %d bytes", len(jsonBody)))
	}

	// Create a new HTTP POST request with context.
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create POST request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request with retry logic and context.
	response, err := executeWithRetry(ctx, request, config)
	if err != nil {
		// Check if error is due to context cancellation
		if errors.Is(err, context.Canceled) {
			log.Warning("⚠️  POST request canceled by context")
			return nil, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("⏰ POST request timed out")
			return nil, err
		}
		// Log and return an error if the request fails.
		log.Error("❌ POST request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}

// Put sends an HTTP PUT request with a JSON body to the specified URL with context support.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Put(url string, body any, headers *Headers) (json.RawMessage, error) {
	return putWithContext(context.Background(), url, body, headers)
}

// putWithContext sends an HTTP PUT request with context support for cancellation.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func putWithContext(ctx context.Context, url string, body any, headers *Headers) (json.RawMessage, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		log.Error("❌ Invalid URL: " + err.Error())
		return nil, err
	}

	// Load configuration
	config := LoadConfig()

	// Log the start of the PUT request.
	log.Info(fmt.Sprintf("📝 Preparing PUT request to %s (Timeout: %v, MaxRetries: %d)",
		url, config.Timeout, config.MaxRetries))

	// Prepare the request body if provided.
	var requestBody io.Reader
	if body != nil {
		// Marshal the body to JSON.
		jsonBody, err := json.Marshal(body)
		if err != nil {
			// Log and return an error if marshaling fails.
			log.Error("❌ Failed to marshal PUT body: " + err.Error())
			return nil, err
		}
		// Create a buffer for the JSON body.
		requestBody = bytes.NewBuffer(jsonBody)

		// Log request body size for debugging
		log.Info(fmt.Sprintf("📦 Request body size: %d bytes", len(jsonBody)))
	}

	// Create a new HTTP PUT request with context.
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, url, requestBody)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create PUT request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request with retry logic and context.
	response, err := executeWithRetry(ctx, request, config)
	if err != nil {
		// Check if error is due to context cancellation
		if errors.Is(err, context.Canceled) {
			log.Warning("⚠️  PUT request canceled by context")
			return nil, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("⏰ PUT request timed out")
			return nil, err
		}
		// Log and return an error if the request fails.
		log.Error("❌ PUT request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}

// Delete sends an HTTP DELETE request to the specified URL with context support.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func Delete(url string, headers *Headers) (json.RawMessage, error) {
	return deleteWithContext(context.Background(), url, headers)
}

// deleteWithContext sends an HTTP DELETE request with context support for cancellation.
// Applies headers and returns the response body as json.RawMessage.
// Returns an error if the request or response processing fails.
func deleteWithContext(ctx context.Context, url string, headers *Headers) (json.RawMessage, error) {
	// Validate URL
	if err := validateURL(url); err != nil {
		log.Error("❌ Invalid URL: " + err.Error())
		return nil, err
	}

	// Load configuration
	config := LoadConfig()

	// Log the start of the DELETE request.
	log.Info(fmt.Sprintf("🗑️  Preparing DELETE request to %s (Timeout: %v, MaxRetries: %d)",
		url, config.Timeout, config.MaxRetries))

	// Create a new HTTP DELETE request with context.
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		// Log and return an error if request creation fails.
		log.Error("❌ Failed to create DELETE request: " + err.Error())
		return nil, err
	}

	// Apply merged headers to the request.
	for headerKey, headerValue := range mergeHeaders(headers) {
		request.Header.Set(headerKey, headerValue)
	}

	// Execute the HTTP request with retry logic and context.
	response, err := executeWithRetry(ctx, request, config)
	if err != nil {
		// Check if error is due to context cancellation
		if errors.Is(err, context.Canceled) {
			log.Warning("⚠️  DELETE request canceled by context")
			return nil, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error("⏰ DELETE request timed out")
			return nil, err
		}
		// Log and return an error if the request fails.
		log.Error("❌ DELETE request failed: " + err.Error())
		return nil, err
	}
	defer response.Body.Close()

	// Process the response and return the result.
	return handleResponse(response)
}
