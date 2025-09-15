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

	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(scanReq)
	if err != nil {
		return nil, fmt.Errorf("can't encode request body: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scan request failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read response body: %w", err)
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
