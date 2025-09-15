package entro

import (
	"errors"
	"fmt"
	"net/http"
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

	return http.DefaultTransport.RoundTrip(req)
}

func NewClient(endpoint string, token string) (client *Client) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryMax = 2
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second
	retryClient.HTTPClient.Timeout = 1 * time.Second
	retryClient.HTTPClient.Transport = &transport{token: token}

	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy

	return &Client{
		endpoint:   endpoint,
		token:      token,
		httpClient: retryClient,
	}
}
