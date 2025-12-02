package entro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/go-retryablehttp"
)

const ScanPrefix = "v2/scan"

type ScanResult struct {
	Origin string `json:"origin"`
	Value  string `json:"value"`
	Line   int    `json:"line"`
}
type ScanResp struct {
	RequestID  string       `json:"requestId"`
	TotalCount int          `json:"totalCount"`
	Results    []ScanResult `json:"results"`
}

type ScanReq struct {
	Data string `json:"data"`
}

func (c *Client) Scan(ctx context.Context, scanReq *ScanReq) (*ScanResp, error) {
	reqURL, err := url.JoinPath(c.endpoint, ScanPrefix)
	if err != nil {
		panic(err)
	}

	// Add generic=true query parameter if enabled
	if os.Getenv("ENTRO_SCAN_GENERICS") == "true" {
		parsedURL, err := url.Parse(reqURL)
		if err != nil {
			return nil, fmt.Errorf("can't parse URL: %w", err)
		}
		query := parsedURL.Query()
		query.Set("generic", "true")
		parsedURL.RawQuery = query.Encode()
		reqURL = parsedURL.String()

		if os.Getenv("ENTRO_DEBUG") == "true" {
			fmt.Println("Debug: Generic scanning enabled")
		}
	}

	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(scanReq)
	if err != nil {
		return nil, fmt.Errorf("can't encode request body: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		log.Fatal(err)
	}

	// Set headers explicitly on the retryablehttp request
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token)

	// Debug: Verify headers are set
	if os.Getenv("ENTRO_DEBUG") == "true" {
		fmt.Printf("Debug: Headers being sent:\n")
		fmt.Printf("  Content-Type: %s\n", req.Header.Get("Content-Type"))
		authHeader := req.Header.Get("Authorization")
		if len(authHeader) > 10 {
			fmt.Printf("  Authorization: %s... (%d chars)\n", authHeader[:10], len(authHeader))
		} else {
			fmt.Printf("  Authorization: %s (%d chars)\n", authHeader, len(authHeader))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scan request failed: %w", err)
	}

	defer resp.Body.Close()

	// Debug: Log response status
	if os.Getenv("ENTRO_DEBUG") == "true" {
		fmt.Printf("Debug: Response status: %d %s\n", resp.StatusCode, resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read response body: %w", err)
		}

		// Log additional details for common error codes
		switch resp.StatusCode {
		case http.StatusForbidden:
			fmt.Fprintf(os.Stderr, "API returned 403 Forbidden - check your API token is valid and has scan permissions\n")
		case http.StatusUnauthorized:
			fmt.Fprintf(os.Stderr, "API returned 401 Unauthorized - check your API token\n")
		case http.StatusTooManyRequests:
			fmt.Fprintf(os.Stderr, "API returned 429 Too Many Requests - you're being rate limited\n")
		}

		return nil, &APIError{
			Code:    resp.StatusCode,
			Message: string(body),
		}
	}

	decoder := json.NewDecoder(resp.Body)
	var result ScanResp
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("can't decode response: %w", err)
	}

	return &result, nil
}
