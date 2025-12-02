# Entro Scan Action

Entro is platform for secrets security and non-human Identity management.
Use entro scan action to scan your pull requests for hardcoded secrets and passwords.

## ⚠️ Important: Checkout Configuration Required

**This action requires proper checkout configuration to work!** By default, GitHub Actions only fetches the merge commit (`fetch-depth: 1`), which means the scanner has no commits to scan.

You **MUST** use the checkout setup shown in the example below, or the action will find zero commits.

## How It Works

The action scans **all changes** in your PR commits:
- ✅ **Additions** - New lines being added to the codebase
- ✅ **Deletions** - Lines being removed from the codebase
- ❌ **Context lines** - Unchanged lines are skipped (they're already in the codebase)

Each commit in the PR is scanned independently, and findings are reported with file location and commit hash.

## Example:

```yaml
name: Scan PR for hardcoded secrets
on:
  pull_request:
    branches: [ "main" ]

jobs:
  secrets-scan:
    runs-on: ubuntu-latest
    steps:
      # Calculate fetch depth to include all PR commits
      - name: 'Get PR commits'
        run: echo "PR_FETCH_DEPTH=$(( ${{ github.event.pull_request.commits }} + 1 ))" >> "${GITHUB_ENV}"

      # Checkout with sufficient depth to scan all commits in the PR
      - name: 'Checkout PR branch and all PR commits'
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.ref }}
          fetch-depth: ${{ env.PR_FETCH_DEPTH }}

      - name: 'Scan for secrets'
        uses: liminal-security/scan-action@v1.0.0
        with:
          api-endpoint: ${{ secrets.API_ENDPOINT }}
          api-token: ${{ secrets.API_KEY }}
```

**Critical Notes:**
- ⚠️ **DO NOT skip the "Get PR commits" step** - it calculates the required fetch depth
- ⚠️ **DO NOT use default checkout** - it will only fetch the merge commit (depth=1)
- ⚠️ **The checkout MUST come before this action** - the action scans what's already checked out

## Troubleshooting

### "No commits found to scan" or "no secrets found" when you know there are secrets

**Problem:** Your workflow is using `fetch-depth: 1` (the default), which only fetches the merge commit.

**Solution:** Add the two-step checkout configuration shown in the example above:
1. Calculate `PR_FETCH_DEPTH` based on the number of commits in the PR
2. Pass it to `actions/checkout` as `fetch-depth: ${{ env.PR_FETCH_DEPTH }}`

**How to verify:** Check your GitHub Actions logs. You should see:
```
Fetching the repository
[command]/usr/bin/git fetch --depth=N origin +COMMIT_SHA:refs/remotes/pull/X/merge
```
Where `N` should equal the number of commits in your PR + 1 (not `--depth=1`).

### Inputs:

1. `api-endpoint` - Entro API endpoint (default: `https://api.entro.security`)
2. `api-token` - Entro API token (create at [settings page](https://app.entro.security/admin/settings?tab=api-keys))
3. `fail-on-error` - Fail workflow on API errors (default: `false`)
   - `false` (default): Log errors and continue scanning other commits
   - `true`: Stop and fail the workflow on first API error
4. `scan-generics` - Scan for generic secrets (default: `false`)
   - `false` (default): Scan for specific secret patterns only
   - `true`: Also scan for generic high-entropy strings and patterns
5. `debug` - Enable debug logging (default: `false`)
   - Logs request details and authorization header info (token length, prefix)
   - Useful for troubleshooting 403 errors or API connectivity issues

### Example with Strict Mode:

If you want the workflow to fail when API errors occur (e.g., network issues, API downtime):

```yaml
- name: 'Scan for secrets'
  uses: liminal-security/scan-action@v1.0.2
  with:
    api-endpoint: ${{ secrets.API_ENDPOINT }}
    api-token: ${{ secrets.API_KEY }}
    fail-on-error: true  # Enable strict mode
```

### Example with Generic Scanning:

To detect generic secrets in addition to specific patterns (AWS keys, GitHub tokens, etc.):

```yaml
- name: 'Scan for secrets'
  uses: liminal-security/scan-action@v1.0.2
  with:
    api-endpoint: ${{ secrets.API_ENDPOINT }}
    api-token: ${{ secrets.API_KEY }}
    scan-generics: true  # Enable generic secret detection
```

**Note:** Generic scanning may find more potential secrets but could also have more false positives.

### Troubleshooting with Debug Mode:

If you're getting 403 errors or want to verify the API token is being sent correctly:

```yaml
- name: 'Scan for secrets'
  uses: liminal-security/scan-action@v1.0.2
  with:
    api-endpoint: ${{ secrets.API_ENDPOINT }}
    api-token: ${{ secrets.API_KEY }}
    debug: true  # Enable debug logging
```

This will show:
- API endpoint being used
- Token length (not the actual token)
- First 10 characters of the token
- Request URLs being called


