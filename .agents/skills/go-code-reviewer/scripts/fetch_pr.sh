#!/bin/bash
# Fetch PR details and prepare for review
# Usage: ./fetch_pr.sh <pr-number-or-url>

set -e

PR_INPUT="$1"

if [ -z "$PR_INPUT" ]; then
    echo "Usage: $0 <pr-number-or-url>"
    exit 1
fi

# Check if gh CLI is available
if ! command -v gh >/dev/null 2>&1; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install with: brew install gh"
    exit 1
fi

# Extract PR number from URL if needed
if [[ "$PR_INPUT" =~ ^https?:// ]]; then
    # Extract PR number from URL like https://github.com/org/repo/pull/123
    PR_NUMBER=$(echo "$PR_INPUT" | grep -oE '[0-9]+$')
else
    PR_NUMBER="$PR_INPUT"
fi

echo "=== Fetching PR #$PR_NUMBER ==="
echo ""

# Get PR details
PR_JSON=$(gh pr view "$PR_NUMBER" --json number,title,author,headRefName,baseRefName,url,body,state)

# Extract details
PR_TITLE=$(echo "$PR_JSON" | jq -r '.title')
PR_AUTHOR=$(echo "$PR_JSON" | jq -r '.author.login')
PR_HEAD=$(echo "$PR_JSON" | jq -r '.headRefName')
PR_BASE=$(echo "$PR_JSON" | jq -r '.baseRefName')
PR_URL=$(echo "$PR_JSON" | jq -r '.url')
PR_BODY=$(echo "$PR_JSON" | jq -r '.body')
PR_STATE=$(echo "$PR_JSON" | jq -r '.state')

echo "PR #$PR_NUMBER: $PR_TITLE"
echo "Author: $PR_AUTHOR"
echo "State: $PR_STATE"
echo "Branch: $PR_HEAD → $PR_BASE"
echo "URL: $PR_URL"
echo ""

# Check if PR is open
if [ "$PR_STATE" != "OPEN" ]; then
    echo "⚠️  Warning: PR is $PR_STATE (not OPEN)"
    echo ""
fi

# Checkout PR branch
echo "Checking out PR branch: $PR_HEAD"
gh pr checkout "$PR_NUMBER"

echo ""
echo "✅ PR #$PR_NUMBER ready for review"
echo ""
echo "Details:"
echo "  Number: $PR_NUMBER"
echo "  Title: $PR_TITLE"
echo "  Author: $PR_AUTHOR"
echo "  Base: $PR_BASE"
echo "  Head: $PR_HEAD"
echo "  URL: $PR_URL"
echo ""

# Export for other scripts to use
echo "$PR_NUMBER" > /tmp/current_pr_number.txt
echo "$PR_BASE" > /tmp/current_pr_base.txt
echo "$PR_URL" > /tmp/current_pr_url.txt

exit 0
