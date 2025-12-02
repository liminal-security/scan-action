package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/liminal-security/scan-action/entro"
	"github.com/liminal-security/scan-action/git"
)

type finding struct {
	fileName string
	origin   string
	value    string
	commit   string
}

func main() { //nolint:funlen
	if len(os.Args) != 2 {
		fmt.Println("Usage: scan-action <git repo>")
		fmt.Println("set ENTRO_API_ENDPOINT and ENTRO_TOKEN environment variables")
		os.Exit(255)
	}

	// Validate configuration
	getEnvVar := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok || val == "" {
			fmt.Printf("Error: %s is not set\n", key)
			fmt.Printf("Add it to your workflow:\n")
			fmt.Printf("  with:\n")
			fmt.Printf("    api-endpoint: ${{ secrets.API_ENDPOINT }}\n")
			fmt.Printf("    api-token: ${{ secrets.API_KEY }}\n")
			os.Exit(255)
		}
		return val
	}

	entroAPIEndpoint := getEnvVar("ENTRO_API_ENDPOINT")
	entroToken := getEnvVar("ENTRO_TOKEN")

	// Validate URL format
	if _, err := url.Parse(entroAPIEndpoint); err != nil {
		fmt.Printf("Error: Invalid API endpoint URL: %s\n", entroAPIEndpoint)
		os.Exit(255)
	}

	// Check if strict mode is enabled
	failOnError := false
	if failOnErrorStr := os.Getenv("ENTRO_FAIL_ON_ERROR"); failOnErrorStr == "true" {
		failOnError = true
		fmt.Println("Strict mode: Will fail on API errors")
	}

	entroClient := entro.NewClient(entroAPIEndpoint, entroToken)

	repoPath := os.Args[1]

	path, err := filepath.Abs(repoPath)
	if err != nil {
		fmt.Printf("can't get absolute path of %s: %s\n", repoPath, err)
		os.Exit(1)
	}

	var findings []finding

	ctx := context.Background()

	differ, err := git.NewDiffer(path)
	if err != nil {
		fmt.Printf("can't create git differ: %s\n", err)
		os.Exit(1)
	}

	commits, err := differ.Diff()
	if err != nil {
		fmt.Printf("can't crate difff: %s\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("Warning: No commits found to scan")
		fmt.Println("Your checkout is too shallow (using fetch-depth: 1)")
		fmt.Println("See: https://github.com/liminal-security/scan-action#example")
		os.Exit(0)
	}

	fmt.Printf("Found %d commit(s) to scan\n", len(commits))

	for _, commit := range commits {
		fmt.Printf("Scanning commit %s\n", commit.Hash)
		r := &entro.ScanReq{
			Data: commit.String(),
		}

		resp, err := entroClient.Scan(ctx, r)
		if err != nil {
			fmt.Printf("Error scanning %s: %s\n", commit.Hash, err)
			if failOnError {
				fmt.Println("Strict mode enabled: Failing due to API error")
				os.Exit(1)
			}
			continue
		}

		if resp.TotalCount > 0 {
			fmt.Printf("Found %d secrets in commit %s\n", resp.TotalCount, commit.Hash)
			for _, res := range resp.Results {
				file, err := commit.GetFileNameByLine(res.Line)
				if err != nil {
					fmt.Printf("error getting file name for line %d: %s\n", res.Line, err)

					continue
				}
				findings = append(findings, finding{
					fileName: file,
					origin:   res.Origin,
					value:    res.Value,
					commit:   commit.Hash,
				})
			}
		} else {
			fmt.Printf("No secrets found in commit %s\n", commit.Hash)
		}
	}

	if len(findings) == 0 {
		fmt.Println("no secrets found")
		os.Exit(0)
	}

	for _, finding := range findings {
		fmt.Printf("::warning file=%s::Found %s: %s in commit %s\n", finding.fileName, finding.origin, finding.value, finding.commit)
	}

	fmt.Printf("Found %d secrets\n", len(findings))
	os.Exit(2)
}
