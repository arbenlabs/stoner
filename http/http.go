package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an HTTP client with additional features
type Client struct {
	httpClient     *http.Client
	baseURL        string
	defaultHeaders map[string]string
	retryConfig    *RetryConfig
	circuitBreaker *CircuitBreaker
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
	Backoff    float64
}

// CircuitBreaker represents circuit breaker configuration
type CircuitBreaker struct {
	MaxFailures  int
	Timeout      time.Duration
	ResetTimeout time.Duration
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// NewClient creates a new HTTP client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:        baseURL,
		defaultHeaders: make(map[string]string),
		retryConfig: &RetryConfig{
			MaxRetries: 3,
			Delay:      1 * time.Second,
			Backoff:    2.0,
		},
		circuitBreaker: &CircuitBreaker{
			MaxFailures:  5,
			Timeout:      30 * time.Second,
			ResetTimeout: 60 * time.Second,
		},
	}
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetRetryConfig sets the retry configuration
func (c *Client) SetRetryConfig(config *RetryConfig) {
	c.retryConfig = config
}

// SetCircuitBreaker sets the circuit breaker configuration
func (c *Client) SetCircuitBreaker(config *CircuitBreaker) {
	c.circuitBreaker = config
}

// SetDefaultHeader sets a default header
func (c *Client) SetDefaultHeader(key, value string) {
	c.defaultHeaders[key] = value
}

// SetDefaultHeaders sets multiple default headers
func (c *Client) SetDefaultHeaders(headers map[string]string) {
	for key, value := range headers {
		c.defaultHeaders[key] = value
	}
}

// Get performs a GET request
func (c *Client) Get(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	})
}

// Post performs a POST request
func (c *Client) Post(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  "POST",
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Put performs a PUT request
func (c *Client) Put(url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  "PUT",
		URL:     url,
		Body:    body,
		Headers: headers,
	})
}

// Delete performs a DELETE request
func (c *Client) Delete(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
	})
}

// Do performs an HTTP request with retry logic
func (c *Client) Do(req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		response, err := c.doRequest(req)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == c.retryConfig.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(c.retryConfig.Delay) *
			pow(c.retryConfig.Backoff, float64(attempt)))

		time.Sleep(delay)
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w",
		c.retryConfig.MaxRetries+1, lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(req *Request) (*Response, error) {
	// Build URL
	url := c.baseURL + req.URL

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range c.defaultHeaders {
		httpReq.Header.Set(key, value)
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type if body is present
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Perform request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// JSON performs a request and unmarshals the response to JSON
func (c *Client) JSON(req *Request, result interface{}) error {
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return json.Unmarshal(resp.Body, result)
}

// GetJSON performs a GET request and unmarshals the response to JSON
func (c *Client) GetJSON(url string, result interface{}, headers map[string]string) error {
	return c.JSON(&Request{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	}, result)
}

// PostJSON performs a POST request and unmarshals the response to JSON
func (c *Client) PostJSON(url string, body interface{}, result interface{}, headers map[string]string) error {
	return c.JSON(&Request{
		Method:  "POST",
		URL:     url,
		Body:    body,
		Headers: headers,
	}, result)
}

// PutJSON performs a PUT request and unmarshals the response to JSON
func (c *Client) PutJSON(url string, body interface{}, result interface{}, headers map[string]string) error {
	return c.JSON(&Request{
		Method:  "PUT",
		URL:     url,
		Body:    body,
		Headers: headers,
	}, result)
}

// DeleteJSON performs a DELETE request and unmarshals the response to JSON
func (c *Client) DeleteJSON(url string, result interface{}, headers map[string]string) error {
	return c.JSON(&Request{
		Method:  "DELETE",
		URL:     url,
		Headers: headers,
	}, result)
}

// Utility functions

// pow calculates x^y
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

// WithContext performs a request with context
func (c *Client) WithContext(ctx context.Context, req *Request) (*Response, error) {
	// Create a copy of the client with context
	clientCopy := *c
	clientCopy.httpClient = &http.Client{
		Timeout: c.httpClient.Timeout,
	}

	// Build URL
	url := c.baseURL + req.URL

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range c.defaultHeaders {
		httpReq.Header.Set(key, value)
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type if body is present
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Perform request
	resp, err := clientCopy.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}

// RateLimiter represents a rate limiter
type RateLimiter struct {
	requests chan time.Time
	rate     time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(chan time.Time, burst),
		rate:     rate,
	}

	// Fill the channel with initial tokens
	for i := 0; i < burst; i++ {
		rl.requests <- time.Now()
	}

	return rl
}

// Wait waits for a token to be available
func (rl *RateLimiter) Wait() {
	now := time.Now()
	select {
	case <-rl.requests:
		// Token available
	default:
		// No token available, wait for one
		<-rl.requests
	}

	// Add new token after rate duration
	go func() {
		time.Sleep(rl.rate)
		rl.requests <- now
	}()
}
