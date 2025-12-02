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

## Error: Missing or Invalid Credentials

### Problem: "unsupported protocol scheme" or similar errors

**Symptoms:**
```
error scanning ...: scan request failed: POST v2/scan giving up after 1 attempt(s): 
Post "v2/scan": unsupported protocol scheme ""
```

### Problem: Configuration validation errors

With the latest version, you'll see detailed validation errors if your configuration is wrong:

```
❌ ERROR: ENTRO_TOKEN environment variable is empty
❌ ERROR: ENTRO_API_ENDPOINT is not a valid URL: not-a-url
❌ ERROR: ENTRO_TOKEN contains template syntax - secret is not being substituted
```

The action now validates:
- ✅ Environment variables exist and aren't empty
- ✅ API endpoint is a valid HTTP/HTTPS URL
- ✅ Token doesn't contain template syntax (means secret wasn't substituted)
- ✅ Token isn't a placeholder value
- ✅ Token has reasonable length

### Solution

Follow the detailed instructions shown in the error output:

**Step 1: Create GitHub Secrets**
1. Go to your repository Settings → Secrets and variables → Actions
2. Create these secrets:
   - Name: `API_ENDPOINT`, Value: `https://api.entro.security`
   - Name: `API_KEY`, Value: your actual Entro API token from [here](https://app.entro.security/admin/settings?tab=api-keys)

**Step 2: Reference Them in Your Workflow**
```yaml
- name: 'Scan for secrets'
  uses: liminal-security/scan-action@v1.0.2
  with:
    api-endpoint: ${{ secrets.API_ENDPOINT }}  # Must match secret name!
    api-token: ${{ secrets.API_KEY }}          # Must match secret name!
```

**Important:** 
- The secret names must match **exactly**. If your secret is named `API_KEY` in GitHub, you must use `${{ secrets.API_KEY }}` in the workflow (not `API_TOKEN`, `ENTRO_TOKEN`, or anything else).
- Secrets must be in **Repository secrets** (not Environment secrets). Environment secrets require additional workflow configuration and won't work by default.

## Error: ENTRO_TOKEN is empty (length: 0)

**Problem:** Debug output shows:
```
ENTRO_TOKEN: [SET] (length: 0)
```

**Common Causes:**

1. **Environment secrets vs Repository secrets** (most common!)
   - If your secret is in **Environment secrets**, the workflow can't access it by default
   - **Solution:** Move the secret to **Repository secrets**:
     - Go to Settings → Secrets and variables → Actions
     - Switch to the **Secrets** tab (not **Environment secrets**)
     - Create the secret there

2. **Secret value is actually empty**
   - The secret exists but has no value
   - **Solution:** Update the secret and paste your actual API token

3. **Whitespace issues**
   - The secret value has trailing newlines or is just spaces
   - **Solution:** Delete and recreate the secret, being careful when pasting

## Still Not Working?

With the latest version, you'll see helpful error messages:

**No commits found:**
```
⚠️  WARNING: No commits found to scan!
This usually means your checkout fetch-depth is too shallow.
```

**Missing credentials:**
```
❌ ERROR: ENTRO_TOKEN environment variable is empty
Make sure your workflow is passing the value correctly...
```

Make sure you're following **all** steps in the example workflow.

