package entro

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClientScan(t *testing.T) {
	tests := []struct {
		name         string
		responseFile string
		responseCode int
		req          *ScanReq
		want         *ScanResp
		wantErr      error
	}{
		{
			name:         "normal one result",
			responseFile: "testdata/one_result.json",
			req: &ScanReq{
				Data: "ghp_BTqLYdZxZZt40adNxffv32CiUiw1R82UC7vz",
			},
			want: &ScanResp{
				RequestID:  "bfdb6eb1-358f-485e-875d-0aff234fab34",
				TotalCount: 1,
				Results: []ScanResult{
					{
						Origin: "GITHUB_API_TOKEN",
						Value:  "ghp_BTqLYdZxZZ************CiUiw1R82UC7vz",
						Line:   1,
					},
				},
			},
		},
		{
			name:         "normal zero results",
			responseFile: "testdata/zero_results.json",
			req: &ScanReq{
				Data: "test",
			},
			want: &ScanResp{
				RequestID:  "7536071d-33b7-4541-a7d2-09b9d3889127",
				TotalCount: 0,
				Results:    []ScanResult{},
			},
		},
		{
			name:         "413 body exceeded 1mb limit",
			responseFile: "testdata/413_to_big",
			responseCode: http.StatusRequestEntityTooLarge,
			req: &ScanReq{
				Data: "Very-very-very-long-string",
			},
			want: nil,

			wantErr: &APIError{Code: http.StatusRequestEntityTooLarge, Message: "Body exceeded 1mb limit"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := "ent_test-token"

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serveFile(t, w, r, tt.responseFile, tt.responseCode, token)
			}))
			defer svr.Close()

			c := NewClient(svr.URL, token)

			got, err := c.Scan(context.Background(), tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Scan() error = %s, wantErr %s", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func testHeaders(t *testing.T, r *http.Request, token string) {
	auth := r.Header.Get("Authorization")
	if auth != token {
		t.Errorf("expected Authorization header to be: %q, got: %q", token, auth)
	}

	accept := r.Header.Get("Content-Type")
	if accept != "application/json" {
		t.Errorf("expected Content-Type header to be: %q, got: %q", "application/json", accept)
	}
}

func serveFile(t *testing.T, w http.ResponseWriter, r *http.Request, name string, code int, token string) {
	testHeaders(t, r, token)

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filePath := filepath.Join(currentDir, name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	if code == 0 {
		code = 200
	}
	w.WriteHeader(code)

	w.Write(data)
}
