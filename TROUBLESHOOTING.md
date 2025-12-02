# Troubleshooting: No Secrets Found

## Problem

Your GitHub Actions log shows:
```
no secrets found
```

But you **know** there are secrets in your PR.

## Root Cause

Your workflow is using the default checkout which only fetches the merge commit (`fetch-depth: 1`). This means:
- ❌ Only the merge commit is fetched
- ❌ Individual PR commits are NOT fetched
- ❌ The scanner has zero commits to scan
- ❌ Result: "no secrets found"

## How to Verify

Check your GitHub Actions log for this line:
```
[command]/usr/bin/git fetch --depth=1 origin +COMMIT_SHA:refs/remotes/pull/X/merge
                            ^^^^^^^^^^
                            This is the problem!
```

If you see `--depth=1`, your workflow is broken.

## Solution

### ❌ WRONG (What you're probably doing now):

```yaml
jobs:
  secrets-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4  # Missing fetch-depth!
      
      - name: 'Scan for secrets'
        uses: liminal-security/scan-action@v1.0.2
        with:
          api-endpoint: ${{ secrets.API_ENDPOINT }}
          api-token: ${{ secrets.API_KEY }}
```

### ✅ CORRECT (What you need to do):

```yaml
jobs:
  secrets-scan:
    runs-on: ubuntu-latest
    steps:
      # STEP 1: Calculate how many commits to fetch
      - name: 'Get PR commits'
        run: echo "PR_FETCH_DEPTH=$(( ${{ github.event.pull_request.commits }} + 1 ))" >> "${GITHUB_ENV}"

      # STEP 2: Checkout with proper depth
      - name: 'Checkout PR branch and all PR commits'
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.ref }}
          fetch-depth: ${{ env.PR_FETCH_DEPTH }}  # This is critical!

      # STEP 3: Now scan
      - name: 'Scan for secrets'
        uses: liminal-security/scan-action@v1.0.2
        with:
          api-endpoint: ${{ secrets.API_ENDPOINT }}
          api-token: ${{ secrets.API_KEY }}
```

## After the Fix

Your logs should now show:
```
Found 2 commit(s) to scan
Scanning commit 12fd1b3...
Found 3 secrets in commit 12fd1b3
::warning file=config.py::Found AWS Access Key: AKIAIOSFODNN7EXAMPLE in commit 12fd1b3
...
```

## Still Not Working?

With the latest version, if no commits are found, you'll see a helpful warning:
```
⚠️  WARNING: No commits found to scan!
This usually means your checkout fetch-depth is too shallow.
```

Make sure you're following **both** checkout steps shown above.

