package main

import (
	"context"
	"fmt"
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

	getEnvVar := func(key string) string {
		val, ok := os.LookupEnv(key)
		if !ok {
			fmt.Printf("%s not set\n", key)
			os.Exit(255)
		} else {
			return val
		}

		return ""
	}

	entroAPIEndpoint := getEnvVar("ENTRO_API_ENDPOINT")
	entroToken := getEnvVar("ENTRO_TOKEN")

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
		fmt.Println("⚠️  WARNING: No commits found to scan!")
		fmt.Println("This usually means your checkout fetch-depth is too shallow.")
		fmt.Println("Please follow the setup instructions in the README:")
		fmt.Println("  https://github.com/liminal-security/scan-action#example")
		fmt.Println()
		fmt.Println("You need to:")
		fmt.Println("  1. Calculate PR commit count: PR_FETCH_DEPTH=$(( ${{ github.event.pull_request.commits }} + 1 ))")
		fmt.Println("  2. Use it in checkout: fetch-depth: ${{ env.PR_FETCH_DEPTH }}")
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
			fmt.Printf("error scanning %s: %s\n", commit.Hash, err)

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
