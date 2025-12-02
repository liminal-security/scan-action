package entro

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP code %d: %s", e.Code, e.Message)
}

func (e *APIError) Is(target error) bool {
	var apiError *APIError

	if errors.As(target, &apiError) {
		if apiError.Code == e.Code {
			return true
		}
	} else {
		return false
	}

	return false
}

type transport struct {
	token string
}

type Client struct {
	endpoint string
	token    string

	httpClient *retryablehttp.Client
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", t.token)

	// Debug logging if enabled
	if os.Getenv("ENTRO_DEBUG") == "true" {
		fmt.Printf("Debug: Sending request to %s\n", req.URL)
		fmt.Printf("Debug: Authorization header length: %d chars\n", len(t.token))
		fmt.Printf("Debug: Authorization header starts with: %s...\n", truncate(t.token, 10))
	}

	return http.DefaultTransport.RoundTrip(req)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Helper to truncate strings - exported for use in other files
func Truncate(s string, maxLen int) string {
	return truncate(s, maxLen)
}

func NewClient(endpoint string, token string) (client *Client) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryMax = 2
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second
	retryClient.HTTPClient.Timeout = 30 * time.Second // Increased from 1s to 30s
	retryClient.HTTPClient.Transport = &transport{token: token}

	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy

	return &Client{
		endpoint:   endpoint,
		token:      token,
		httpClient: retryClient,
	}
}
