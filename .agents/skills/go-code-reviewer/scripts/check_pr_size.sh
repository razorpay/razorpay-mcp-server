#!/bin/bash
# Check PR size metrics against Layer 2 constraints with GitHub API verification
# Usage: ./check_pr_size.sh [pr-number] [base-branch]
# Prioritizes GitHub API (source of truth) over git diff (local state)
#
# Features:
# - Uses GitHub API as primary source (gh pr view)
# - Verifies git diff matches GitHub API
# - Warns on mismatches (local state ≠ PR state)
# - Falls back gracefully when PR number unavailable

set -e

PR_NUMBER="${1}"
BASE_BRANCH="${2}"

# Check if we have a PR number from fetch_pr.sh
if [ -z "$PR_NUMBER" ] && [ -f /tmp/current_pr_number.txt ]; then
    PR_NUMBER=$(cat /tmp/current_pr_number.txt)
fi

# Check if we have a PR base branch from fetch_pr.sh
if [ -z "$BASE_BRANCH" ] && [ -f /tmp/current_pr_base.txt ]; then
    BASE_BRANCH=$(cat /tmp/current_pr_base.txt)
else
    BASE_BRANCH="${BASE_BRANCH:-main}"
    # Try main first, fall back to master
    if ! git rev-parse --verify "$BASE_BRANCH" >/dev/null 2>&1; then
        if git rev-parse --verify master >/dev/null 2>&1; then
            BASE_BRANCH="master"
        fi
    fi
fi

echo "=== Layer 2: Scope Boundary Analysis ==="
echo "Comparing against: $BASE_BRANCH"
echo ""

# Try GitHub API first (SOURCE OF TRUTH)
if [ -n "$PR_NUMBER" ] && command -v gh &> /dev/null; then
    echo "📊 Fetching metrics from GitHub API (source of truth)..."

    # Get PR metrics from GitHub
    GH_FILES=$(gh pr view "$PR_NUMBER" --json files --jq '.files | length' 2>/dev/null || echo "")
    GH_ADDITIONS=$(gh pr view "$PR_NUMBER" --json additions --jq '.additions' 2>/dev/null || echo "")
    GH_DELETIONS=$(gh pr view "$PR_NUMBER" --json deletions --jq '.deletions' 2>/dev/null || echo "")

    if [ -n "$GH_FILES" ] && [ -n "$GH_ADDITIONS" ] && [ -n "$GH_DELETIONS" ]; then
        GH_TOTAL_LINES=$((GH_ADDITIONS + GH_DELETIONS))

        echo "✅ GitHub API metrics retrieved"
        echo ""
        echo "GitHub PR #$PR_NUMBER (Source of Truth):"
        echo "  Files changed:     $GH_FILES"
        echo "  Lines changed:     $GH_TOTAL_LINES ($GH_ADDITIONS additions, $GH_DELETIONS deletions)"

        # Get file list from GitHub for type analysis
        GH_FILES_LIST=$(gh pr view "$PR_NUMBER" --json files --jq '.files[].path' 2>/dev/null)
        GH_GO_FILES=$(echo "$GH_FILES_LIST" | grep '\.go$' | wc -l | tr -d ' ' || echo "0")
        GH_NON_GO_FILES=$(echo "$GH_FILES_LIST" | grep -v '\.go$' | grep -v '^$' | wc -l | tr -d ' ' || echo "0")

        echo "  Go files:          $GH_GO_FILES"
        echo "  Non-Go files:      $GH_NON_GO_FILES"
        echo ""

        # Now get git diff for verification
        echo "🔍 Verifying against local git diff..."
        GIT_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null | wc -l | tr -d ' ' || echo "0")
        GIT_LINES=$(git diff --shortstat "$BASE_BRANCH"...HEAD 2>/dev/null | awk '{print $4+$6}' || echo "0")

        echo "Local git diff (for verification):"
        echo "  Files changed:     $GIT_FILES"
        echo "  Lines changed:     $GIT_LINES"
        echo ""

        # Verification check
        if [ "$GH_FILES" -ne "$GIT_FILES" ] || [ "$GH_TOTAL_LINES" -ne "$GIT_LINES" ]; then
            echo "⚠️  MISMATCH DETECTED between GitHub API and local git diff"
            echo ""
            echo "Discrepancy:"
            echo "  Files:  GitHub=$GH_FILES, Local=$GIT_FILES (diff: $((GIT_FILES - GH_FILES)))"
            echo "  Lines:  GitHub=$GH_TOTAL_LINES, Local=$GIT_LINES (diff: $((GIT_LINES - GH_TOTAL_LINES)))"
            echo ""
            echo "Possible causes:"
            echo "  • Uncommitted local changes"
            echo "  • Local branch has merged other changes"
            echo "  • Local master/main is stale"
            echo ""
            echo "✅ Using GitHub API as source of truth for review"

            # Show which files are different
            if [ "$GH_FILES" -ne "$GIT_FILES" ]; then
                echo ""
                echo "Files in local diff but NOT in GitHub PR:"
                comm -13 \
                    <(echo "$GH_FILES_LIST" | sort) \
                    <(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null | sort) \
                    | head -10
            fi
            echo ""
        else
            echo "✅ VERIFIED: GitHub API and git diff match"
            echo ""
        fi

        # Use GitHub metrics for scoring
        FILES_CHANGED=$GH_FILES
        LINES_CHANGED=$GH_TOTAL_LINES
        GO_FILES_CHANGED=$GH_GO_FILES
        NON_GO_FILES=$GH_NON_GO_FILES
        CHANGED_FILES="$GH_FILES_LIST"
        SOURCE="GitHub API"

    else
        echo "⚠️  GitHub API unavailable, falling back to git diff"
        echo ""
        SOURCE="git diff (fallback)"
        # Fall through to git diff method below
    fi
fi

# Fallback to git diff if GitHub API unavailable
if [ -z "$FILES_CHANGED" ]; then
    echo "📊 Using git diff (GitHub API unavailable)"
    echo ""

    # Get diff stats
    LINES_CHANGED=$(git diff --shortstat "$BASE_BRANCH"...HEAD 2>/dev/null | awk '{print $4+$6}' || echo "0")
    FILES_CHANGED=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null | wc -l | tr -d ' ' || echo "0")
    CHANGED_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null || echo "")
    GO_FILES_CHANGED=$(echo "$CHANGED_FILES" | grep '\.go$' | wc -l | tr -d ' ' || echo "0")
    NON_GO_FILES=$(echo "$CHANGED_FILES" | grep -v '\.go$' | grep -v '^$' | wc -l | tr -d ' ' || echo "0")
    SOURCE="git diff"

    echo "⚠️  WARNING: Using local git state, may include uncommitted changes"
    echo "   For accurate metrics, provide PR number: ./check_pr_size_v2.sh <pr-number>"
    echo ""
fi

# Determine status based on thresholds
STATUS="✅ PASS"
WARNING=""

if [ "$LINES_CHANGED" -gt 400 ]; then
    STATUS="❌ FAIL"
    WARNING="Lines changed ($LINES_CHANGED) exceeds hard limit (400)"
elif [ "$LINES_CHANGED" -gt 200 ]; then
    STATUS="⚠️  WARNING"
    WARNING="Lines changed ($LINES_CHANGED) exceeds target (200)"
fi

if [ "$FILES_CHANGED" -gt 15 ]; then
    STATUS="❌ FAIL"
    WARNING="$WARNING\nFiles changed ($FILES_CHANGED) exceeds limit (15)"
elif [ "$FILES_CHANGED" -gt 10 ]; then
    if [ "$STATUS" != "❌ FAIL" ]; then
        STATUS="⚠️  WARNING"
    fi
    WARNING="$WARNING\nFiles changed ($FILES_CHANGED) exceeds target (10)"
fi

# Display summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Status: $STATUS"
echo "Source: $SOURCE"
echo ""
echo "Metrics:"
echo "  Lines changed:     $LINES_CHANGED (target: <200, limit: <400)"
echo "  Files changed:     $FILES_CHANGED (target: <10, limit: <15)"
echo "  Go files:          $GO_FILES_CHANGED"
echo "  Non-Go files:      $NON_GO_FILES"
echo ""

if [ -n "$WARNING" ]; then
    echo "⚠️  Warnings:"
    echo -e "$WARNING"
    echo ""
    echo "Recommendation: Consider splitting this PR into smaller, focused changes."
fi

echo ""
echo "Changed files by type:"
echo "$CHANGED_FILES" | sed 's/^/  /'
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Exit with appropriate code
if [ "$STATUS" = "❌ FAIL" ]; then
    exit 1
elif [ "$STATUS" = "⚠️  WARNING" ]; then
    exit 2
else
    exit 0
fi
