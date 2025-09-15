# Entro Scan Action

Entro is platform for secrets security and non-human Identity management.
Use entro scan action to scan your pull requests for hardcoded secrets and passwords.

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
      - name: 'Get PR commits'
        run: echo "PR_FETCH_DEPTH=$(( ${{ github.event.pull_request.commits }} + 1 ))" >> "${GITHUB_ENV}"

      - name: 'Checkout PR branch and all PR commits'
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.ref }}
          fetch-depth: ${{ env.PR_FETCH_DEPTH }}

      - name: 'Scan for secrets'
        uses: liminal-security/scan-action@v1
        with:
          api-endpoint: ${{ secrets.API_ENDPOINT }}
          api-token: ${{ secrets.API_KEY }}
```

### Inputs:

1. `api-endpoint` Entro API endpoint usually `https://api.entro.security`
2. `api-token` Entro API token can be created at [settings page](https://app.entro.security/admin/settings?tab=api-keys).


