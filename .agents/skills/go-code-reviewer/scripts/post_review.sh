#!/bin/bash
# Post review results to GitHub PR with inline comments support
# Usage: ./post_review.sh <pr-number> <review-file> <action> [inline-comments-file]
# Actions: approve, request-changes, comment
# Inline comments file format: JSON array of {path, line, body} objects

set -e

PR_NUMBER="$1"
REVIEW_FILE="$2"
ACTION="$3"
INLINE_COMMENTS_FILE="$4"

# Validation
if [ -z "$PR_NUMBER" ] || [ -z "$REVIEW_FILE" ] || [ -z "$ACTION" ]; then
    echo "Usage: $0 <pr-number> <review-file> <action> [inline-comments-file]"
    echo "Actions: approve, request-changes, comment"
    echo ""
    echo "Inline comments file (optional): JSON file with format:"
    echo '  [{"path": "file.go", "line": 42, "body": "Comment text"}, ...]'
    exit 1
fi

if [ ! -f "$REVIEW_FILE" ]; then
    echo "Error: Review file not found: $REVIEW_FILE"
    exit 1
fi

# Pre-posting validation
echo "=== Validating Review ==="

# Check if gh CLI is available
if ! command -v gh >/dev/null 2>&1; then
    echo "❌ GitHub CLI (gh) is not installed"
    echo "Install with: brew install gh"
    exit 1
fi
echo "✅ GitHub CLI found"

# Check if jq is available (needed for JSON processing)
if ! command -v jq >/dev/null 2>&1; then
    echo "❌ jq is not installed (required for inline comments)"
    echo "Install with: brew install jq"
    exit 1
fi
echo "✅ jq found"

# Check if authenticated
if ! gh auth status >/dev/null 2>&1; then
    echo "❌ Not authenticated with GitHub"
    echo "Run: gh auth login"
    exit 1
fi
echo "✅ GitHub authenticated"

# Validate review file is valid markdown
if ! head -1 "$REVIEW_FILE" >/dev/null 2>&1; then
    echo "❌ Review file is empty or unreadable"
    exit 1
fi
echo "✅ Review file valid"

# Validate inline comments JSON if provided
if [ -n "$INLINE_COMMENTS_FILE" ] && [ -f "$INLINE_COMMENTS_FILE" ]; then
    if ! jq empty "$INLINE_COMMENTS_FILE" 2>/dev/null; then
        echo "❌ Inline comments file is not valid JSON"
        exit 1
    fi
    echo "✅ Inline comments JSON valid"
fi

# Validate PR exists
if ! gh pr view "$PR_NUMBER" >/dev/null 2>&1; then
    echo "❌ PR #$PR_NUMBER not found"
    exit 1
fi
echo "✅ PR #$PR_NUMBER found"

echo ""
echo "=== Posting Review to PR #$PR_NUMBER ==="
echo "Action: $ACTION"
echo ""

# Read review content
REVIEW_BODY=$(cat "$REVIEW_FILE")

# Map action to GitHub API event
case "$ACTION" in
    approve)
        EVENT="APPROVE"
        echo "📝 Posting APPROVAL..."
        ;;
    request-changes)
        EVENT="REQUEST_CHANGES"
        echo "📝 Posting REQUEST CHANGES..."
        ;;
    comment)
        EVENT="COMMENT"
        echo "📝 Posting COMMENT..."
        ;;
    *)
        echo "❌ Invalid action: $ACTION"
        echo "Valid actions: approve, request-changes, comment"
        exit 1
        ;;
esac

# Process inline comments
if [ -n "$INLINE_COMMENTS_FILE" ] && [ -f "$INLINE_COMMENTS_FILE" ]; then
    echo "Processing inline comments from: $INLINE_COMMENTS_FILE"

    # Count inline comments
    COMMENT_COUNT=$(jq 'length' "$INLINE_COMMENTS_FILE")
    echo "  → $COMMENT_COUNT inline comment(s) found"

    # Get PR details
    PR_HEAD_SHA=$(gh pr view "$PR_NUMBER" --json headRefOid -q .headRefOid)
    echo "  → PR commit: ${PR_HEAD_SHA:0:7}"

    # Get list of changed files
    CHANGED_FILES=$(gh pr view "$PR_NUMBER" --json files --jq '.files[].path')

    # Transform inline comments to handle new files vs modified files
    echo "  → Transforming comments for GitHub API..."

    TRANSFORMED_COMMENTS=$(jq -r --arg changed_files "$CHANGED_FILES" '
        map({
            path: .path,
            body: .body,
            line: (.line // .position),
            original_line: .line
        })
    ' "$INLINE_COMMENTS_FILE")

    # Save transformed comments for debugging
    echo "$TRANSFORMED_COMMENTS" > /tmp/pr${PR_NUMBER}_transformed_comments.json

    # Build review request with inline comments
    REVIEW_JSON=$(jq -n \
        --arg body "$REVIEW_BODY" \
        --arg event "$EVENT" \
        --arg commit_id "$PR_HEAD_SHA" \
        --argjson comments "$TRANSFORMED_COMMENTS" \
        '{
            body: $body,
            event: $event,
            commit_id: $commit_id,
            comments: $comments
        }')

    echo "  → Posting review with inline comments..."

    # Save request for debugging
    echo "$REVIEW_JSON" > /tmp/pr${PR_NUMBER}_review_request.json

    # Post review using GitHub API with error handling
    if gh api \
        --method POST \
        -H "Accept: application/vnd.github+json" \
        "/repos/{owner}/{repo}/pulls/$PR_NUMBER/reviews" \
        --input - <<< "$REVIEW_JSON" > /tmp/pr${PR_NUMBER}_response.json 2>&1; then
        echo "✅ Review with inline comments posted successfully!"
    else
        echo "⚠️  Inline comments failed. Falling back to review body only..."
        echo "Error details saved to: /tmp/pr${PR_NUMBER}_response.json"
        echo ""

        # Fallback: post without inline comments
        case "$ACTION" in
            approve)
                gh pr review "$PR_NUMBER" --approve --body "$REVIEW_BODY"
                ;;
            request-changes)
                gh pr review "$PR_NUMBER" --request-changes --body "$REVIEW_BODY"
                ;;
            comment)
                gh pr review "$PR_NUMBER" --comment --body "$REVIEW_BODY"
                ;;
        esac

        echo "✅ Review body posted (without inline comments)"
        echo ""
        echo "📝 To add inline comments manually:"
        cat "$INLINE_COMMENTS_FILE" | jq -r '.[] | "  • \(.path):\(.line) - \(.body | split("\n")[0])"'
    fi

else
    echo "No inline comments file provided - posting review body only"
    echo ""

    # Post review without inline comments
    case "$ACTION" in
        approve)
            gh pr review "$PR_NUMBER" --approve --body "$REVIEW_BODY"
            ;;
        request-changes)
            gh pr review "$PR_NUMBER" --request-changes --body "$REVIEW_BODY"
            ;;
        comment)
            gh pr review "$PR_NUMBER" --comment --body "$REVIEW_BODY"
            ;;
    esac

    echo "✅ Review posted successfully!"
fi

echo ""
echo "View at: $(gh pr view "$PR_NUMBER" --json url -q .url)"
echo ""
echo "Review artifacts saved:"
echo "  • Request: /tmp/pr${PR_NUMBER}_review_request.json"
echo "  • Comments: /tmp/pr${PR_NUMBER}_transformed_comments.json"
echo "  • Response: /tmp/pr${PR_NUMBER}_response.json"
