#!/bin/bash
# Generate template inline comments JSON from git diff
# Usage: ./generate_inline_comments_template.sh [base-branch] [output-file]
#
# This script creates a JSON template with all changed lines, making it easy
# to add inline comments during Layer 3 review

set -e

BASE_BRANCH="${1:-main}"
OUTPUT_FILE="${2:-/tmp/inline_comments_template.json}"

echo "=== Generating Inline Comments Template ==="
echo ""
echo "Base branch: $BASE_BRANCH"
echo "Output file: $OUTPUT_FILE"
echo ""

# Get list of changed files
CHANGED_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null)

if [ -z "$CHANGED_FILES" ]; then
    echo "Error: No changes detected"
    exit 1
fi

# Start JSON array
echo "[" > "$OUTPUT_FILE"

FIRST_ENTRY=true

# For each changed file, extract changed line numbers
while IFS= read -r file; do
    if [ -z "$file" ]; then
        continue
    fi

    # Get line numbers that changed (only additions and modifications)
    # Use sed instead of awk for better portability
    CHANGED_LINES=$(git diff --unified=0 "$BASE_BRANCH"...HEAD -- "$file" | \
        grep '^@@' | \
        sed -E 's/^@@ -[0-9]+(,[0-9]+)? \+([0-9]+)(,[0-9]+)? @@.*/\2 \3/' | \
        while read start count; do
            # Remove leading comma if present
            count=${count#,}
            # Default to 1 if count is empty
            count=${count:-1}
            # Print each line number in the range
            seq "$start" $((start + count - 1))
        done)

    if [ -z "$CHANGED_LINES" ]; then
        continue
    fi

    # Generate JSON entries for significant lines (every 10th line as examples)
    LINE_NUM=0
    for line in $CHANGED_LINES; do
        LINE_NUM=$((LINE_NUM + 1))

        # Only include every 10th line to keep template manageable
        if [ $((LINE_NUM % 10)) -ne 0 ] && [ $LINE_NUM -ne 1 ]; then
            continue
        fi

        # Add comma before entry if not first
        if [ "$FIRST_ENTRY" = false ]; then
            echo "," >> "$OUTPUT_FILE"
        fi
        FIRST_ENTRY=false

        # Get the actual code at this line for context
        # Escape JSON special chars and remove control chars for valid JSON
        CODE_LINE=$(sed -n "${line}p" "$file" 2>/dev/null | \
            tr -d '\000-\037' | \
            sed 's/\\/\\\\/g; s/"/\\"/g' | \
            head -c 60)

        cat >> "$OUTPUT_FILE" <<EOF
  {
    "path": "$file",
    "line": $line,
    "body": "TODO: Add comment for: $CODE_LINE..."
  }
EOF
    done

done <<< "$CHANGED_FILES"

# Close JSON array
echo "" >> "$OUTPUT_FILE"
echo "]" >> "$OUTPUT_FILE"

# Count generated entries
ENTRY_COUNT=$(grep -c '"path"' "$OUTPUT_FILE" || echo "0")

echo "✅ Generated template with $ENTRY_COUNT placeholder entries"
echo ""
echo "Template saved to: $OUTPUT_FILE"
echo ""
echo "Next steps:"
echo "  1. Open $OUTPUT_FILE in your editor"
echo "  2. Replace 'TODO: Add comment' with actual review comments"
echo "  3. Remove entries you don't need"
echo "  4. Add more entries as needed with format:"
echo "     {\"path\": \"file.go\", \"line\": 42, \"body\": \"Your comment\"}"
echo "  5. Use with: ./post_review_with_comments_v2.sh <pr> <review> comment $OUTPUT_FILE"
echo ""
echo "💡 Tip: Focus on significant issues - not every changed line needs a comment!"
echo ""

exit 0
