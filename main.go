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

	// Debug: Show all relevant environment variables
	if os.Getenv("ENTRO_DEBUG") == "true" {
		fmt.Println("Debug: Environment variables:")
		for _, env := range []string{"ENTRO_API_ENDPOINT", "ENTRO_TOKEN", "ENTRO_FAIL_ON_ERROR", "ENTRO_DEBUG"} {
			val, exists := os.LookupEnv(env)
			if exists {
				if env == "ENTRO_TOKEN" {
					fmt.Printf("  %s: [SET] (length: %d)\n", env, len(val))
				} else {
					fmt.Printf("  %s: %s\n", env, val)
				}
			} else {
				fmt.Printf("  %s: [NOT SET]\n", env)
			}
		}
		fmt.Println()
	}

	// Validate configuration
	getEnvVar := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			fmt.Printf("Error: %s environment variable is not set\n", key)
			fmt.Printf("This means the action.yml is not passing it correctly.\n")
			os.Exit(255)
		}
		if val == "" {
			fmt.Printf("Error: %s is empty\n", key)
			fmt.Printf("Your GitHub secret exists but has no value, or the secret name doesn't match.\n")
			fmt.Printf("\nCheck:\n")
			fmt.Printf("  1. Secret exists in: Settings → Secrets and variables → Actions\n")
			fmt.Printf("  2. Secret has a value (not empty)\n")
			fmt.Printf("  3. Secret name in workflow matches exactly (case-sensitive)\n")
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

	// Show token info (length only, not the actual token)
	fmt.Printf("API Endpoint: %s\n", entroAPIEndpoint)
	fmt.Printf("Token configured: yes (%d characters)\n", len(entroToken))

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
