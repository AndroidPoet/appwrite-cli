package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AndroidPoet/appwrite-cli/internal/config"
)

// Client wraps the Appwrite API client
type Client struct {
	httpClient *http.Client
	apiKey     string
	projectID  string
	endpoint   string
	timeout    time.Duration
}

// APIError represents a structured error from the Appwrite API
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Code       int    `json:"code,omitempty"`
	Type       string `json:"type,omitempty"`
	Version    string `json:"version,omitempty"`
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("%s: %s", e.Type, e.Message)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// debugTransport wraps http.RoundTripper to log requests
type debugTransport struct {
	base http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("DEBUG: %s %s\n", req.Method, req.URL)
	for k, v := range req.Header {
		if k == "X-Appwrite-Key" {
			val := v[0]
			if len(val) > 10 {
				fmt.Printf("DEBUG:   %s: %s...%s\n", k, val[:6], val[len(val)-4:])
			}
		} else {
			fmt.Printf("DEBUG:   %s: %s\n", k, v)
		}
	}
	resp, err := t.base.RoundTrip(req)
	if err == nil {
		fmt.Printf("DEBUG: Response: %d %s\n", resp.StatusCode, resp.Status)
	}
	return resp, err
}

// NewClient creates a new API client
func NewClient(projectID string, timeout time.Duration) (*Client, error) {
	apiKey, err := config.GetAPIKey()
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport
	if config.IsDebug() {
		transport = &debugTransport{base: transport}
	}

	return &Client{
		httpClient: &http.Client{Transport: transport},
		apiKey:     apiKey,
		projectID:  projectID,
		endpoint:   config.GetEndpoint(),
		timeout:    timeout,
	}, nil
}

// GetProjectID returns the project ID
func (c *Client) GetProjectID() string {
	return c.projectID
}

// GetEndpoint returns the API endpoint
func (c *Client) GetEndpoint() string {
	return c.endpoint
}

// Context returns a context with timeout
func (c *Client) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

// Do executes an API request
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.endpoint + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Appwrite uses custom headers for auth
	req.Header.Set("X-Appwrite-Project", c.projectID)
	req.Header.Set("X-Appwrite-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(respBody, apiErr); err != nil {
			apiErr.Message = string(respBody)
		}
		return apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.Do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, http.MethodPut, path, body, result)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}

// ListResponse represents a paginated list response from Appwrite
type ListResponse struct {
	Total    int             `json:"total"`
	Items    json.RawMessage `json:"-"`
}

// ListAll fetches all pages using offset-based pagination
func (c *Client) ListAll(ctx context.Context, path string, limit int, itemsKey string, collector func(json.RawMessage) error) error {
	offset := 0
	for {
		pagePath := fmt.Sprintf("%s?limit=%d&offset=%d", path, limit, offset)

		var raw json.RawMessage
		if err := c.Get(ctx, pagePath, &raw); err != nil {
			return err
		}

		// Parse the response to get total and items
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return fmt.Errorf("failed to parse list response: %w", err)
		}

		items, ok := envelope[itemsKey]
		if !ok {
			return fmt.Errorf("response missing '%s' field", itemsKey)
		}

		if err := collector(items); err != nil {
			return err
		}

		var total struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(raw, &total); err != nil {
			return err
		}

		offset += limit
		if offset >= total.Total {
			break
		}
	}
	return nil
}
