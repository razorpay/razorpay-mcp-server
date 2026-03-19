#!/bin/bash
# Track review history for PRs to show changes between reviews
# Usage: ./track_review_history.sh <pr-number> <review-file>

set -e

PR_NUMBER="$1"
REVIEW_FILE="$2"

if [ -z "$PR_NUMBER" ] || [ -z "$REVIEW_FILE" ]; then
    echo "Usage: $0 <pr-number> <review-file>"
    exit 1
fi

HISTORY_DIR="/tmp/go_code_reviewer_history"
PR_HISTORY_DIR="$HISTORY_DIR/pr${PR_NUMBER}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create history directory
mkdir -p "$PR_HISTORY_DIR"

# Save current review
cp "$REVIEW_FILE" "$PR_HISTORY_DIR/review_${TIMESTAMP}.md"

# Get PR stats
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
    gh pr view "$PR_NUMBER" --json number,title,headRefOid,additions,deletions,changedFiles > \
        "$PR_HISTORY_DIR/stats_${TIMESTAMP}.json" 2>/dev/null || true
fi

# Find previous review
PREVIOUS_REVIEW=$(ls -t "$PR_HISTORY_DIR"/review_*.md 2>/dev/null | sed -n '2p')

if [ -n "$PREVIOUS_REVIEW" ]; then
    echo "=== Review History for PR #$PR_NUMBER ==="
    echo ""
    echo "Previous review: $(basename "$PREVIOUS_REVIEW")"
    echo "Current review:  review_${TIMESTAMP}.md"
    echo ""

    # Extract key metrics from reviews if possible
    PREV_TIMESTAMP=$(basename "$PREVIOUS_REVIEW" | sed 's/review_//;s/.md//')

    echo "Changes detected:"

    # Compare file sizes
    PREV_SIZE=$(wc -l < "$PREVIOUS_REVIEW")
    CURR_SIZE=$(wc -l < "$REVIEW_FILE")
    SIZE_DIFF=$((CURR_SIZE - PREV_SIZE))

    if [ $SIZE_DIFF -gt 0 ]; then
        echo "  • Review length: +${SIZE_DIFF} lines (more detailed)"
    elif [ $SIZE_DIFF -lt 0 ]; then
        echo "  • Review length: ${SIZE_DIFF} lines (more concise)"
    else
        echo "  • Review length: unchanged"
    fi

    # Try to extract LOC metrics from both reviews
    PREV_LOC=$(grep -o "Lines changed:.*[0-9]\+" "$PREVIOUS_REVIEW" | grep -o "[0-9]\+" | head -1)
    CURR_LOC=$(grep -o "Lines changed:.*[0-9]\+" "$REVIEW_FILE" | grep -o "[0-9]\+" | head -1)

    if [ -n "$PREV_LOC" ] && [ -n "$CURR_LOC" ]; then
        LOC_DIFF=$((CURR_LOC - PREV_LOC))
        if [ $LOC_DIFF -ne 0 ]; then
            if [ $LOC_DIFF -gt 0 ]; then
                echo "  • PR size: +${LOC_DIFF} LOC (PR got larger)"
            else
                echo "  • PR size: ${LOC_DIFF} LOC (PR got smaller)"
            fi
        fi
    fi

    # Check for fixed issues
    if grep -q "✅.*FIXED" "$REVIEW_FILE" 2>/dev/null; then
        echo "  • Issues fixed:"
        grep "✅.*FIXED" "$REVIEW_FILE" | sed 's/^/    - /' | head -3
    fi

    # Check for new issues
    if grep -q "❌.*NEW\|⚠️.*NEW" "$REVIEW_FILE" 2>/dev/null; then
        echo "  • New issues found:"
        grep "❌.*NEW\|⚠️.*NEW" "$REVIEW_FILE" | sed 's/^/    - /' | head -3
    fi

    echo ""
    echo "Review history saved to: $PR_HISTORY_DIR"
    echo "  • Current:  $PR_HISTORY_DIR/review_${TIMESTAMP}.md"
    echo "  • Previous: $PREVIOUS_REVIEW"

    # Create comparison file
    COMPARISON_FILE="$PR_HISTORY_DIR/comparison_${PREV_TIMESTAMP}_to_${TIMESTAMP}.md"
    cat > "$COMPARISON_FILE" <<EOF
# Review Comparison: PR #$PR_NUMBER

## Timeline
- Previous review: ${PREV_TIMESTAMP}
- Current review: ${TIMESTAMP}
- Time between reviews: $((($(date -j -f "%Y%m%d_%H%M%S" "${TIMESTAMP}" +%s) - $(date -j -f "%Y%m%d_%H%M%S" "${PREV_TIMESTAMP}" +%s)) / 3600)) hours

## Previous Review
Location: $PREVIOUS_REVIEW
Size: $PREV_SIZE lines

## Current Review
Location: $PR_HISTORY_DIR/review_${TIMESTAMP}.md
Size: $CURR_SIZE lines

## Changes
- Review length: $SIZE_DIFF lines
- PR size: ${LOC_DIFF:-unknown} LOC change

## Full Reviews
See individual review files for complete details.
EOF

    echo "  • Comparison: $COMPARISON_FILE"

else
    echo "=== First Review for PR #$PR_NUMBER ==="
    echo ""
    echo "Review saved to: $PR_HISTORY_DIR/review_${TIMESTAMP}.md"
    echo "This is the first review - no comparison available"
fi

echo ""
echo "History location: $PR_HISTORY_DIR"
echo "Total reviews: $(ls -1 "$PR_HISTORY_DIR"/review_*.md 2>/dev/null | wc -l | tr -d ' ')"
