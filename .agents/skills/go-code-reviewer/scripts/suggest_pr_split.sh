#!/bin/bash
# Suggest how to split a large PR into smaller, reviewable PRs with exact git commands
# Usage: ./suggest_pr_split.sh [base-branch]

set -e

BASE_BRANCH="${1:-main}"

# Get PR metrics
TOTAL_LOC=$(git diff --shortstat "$BASE_BRANCH"...HEAD 2>/dev/null | awk '{print $4+$6}')
TOTAL_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null | wc -l | tr -d ' ')

if [ "$TOTAL_LOC" -lt 400 ]; then
    echo "✅ PR size is acceptable ($TOTAL_LOC LOC < 400 LOC threshold)"
    echo "No splitting needed."
    exit 0
fi

echo "=== PR Splitting Recommendations ==="
echo ""
echo "Current PR size: $TOTAL_LOC LOC across $TOTAL_FILES files"
echo "Target: <200 LOC per PR (hard limit: 400 LOC)"
echo ""

# Get all changed files with their line counts
CHANGED_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD)

echo "Analyzing changed files..."
echo ""

# Categorize files by type/package
declare -A FILE_CATEGORIES

while IFS= read -r file; do
    if [ -z "$file" ]; then
        continue
    fi

    # Count lines changed in this file
    LINES=$(git diff --shortstat "$BASE_BRANCH"...HEAD -- "$file" 2>/dev/null | awk '{print $4+$6}')

    # Categorize by directory/package
    DIR=$(dirname "$file")
    PKG_NAME=$(basename "$DIR")

    # Special categories
    if [[ "$file" == *_test.go ]]; then
        CATEGORY="tests"
    elif [[ "$file" == go.mod ]] || [[ "$file" == go.sum ]]; then
        CATEGORY="dependencies"
    elif [[ "$file" == *.proto ]]; then
        CATEGORY="protobuf"
    elif [[ "$file" == *.md ]]; then
        CATEGORY="documentation"
    elif [[ "$DIR" == *migration* ]] || [[ "$DIR" == *schema* ]]; then
        CATEGORY="database"
    elif [[ "$DIR" == *handler* ]] || [[ "$DIR" == *api* ]]; then
        CATEGORY="api"
    else
        CATEGORY="$PKG_NAME"
    fi

    # Add file to category
    if [ -z "${FILE_CATEGORIES[$CATEGORY]}" ]; then
        FILE_CATEGORIES[$CATEGORY]="$file:$LINES"
    else
        FILE_CATEGORIES[$CATEGORY]="${FILE_CATEGORIES[$CATEGORY]}|$file:$LINES"
    fi
done <<< "$CHANGED_FILES"

# Calculate total LOC per category
declare -A CATEGORY_LOC

for category in "${!FILE_CATEGORIES[@]}"; do
    TOTAL=0
    IFS='|' read -ra FILES <<< "${FILE_CATEGORIES[$category]}"
    for file_entry in "${FILES[@]}"; do
        LINES=$(echo "$file_entry" | cut -d: -f2)
        TOTAL=$((TOTAL + LINES))
    done
    CATEGORY_LOC[$category]=$TOTAL
done

# Sort categories by size (largest first)
SORTED_CATEGORIES=($(
    for category in "${!CATEGORY_LOC[@]}"; do
        echo "${CATEGORY_LOC[$category]}:$category"
    done | sort -rn | cut -d: -f2
))

echo "File categories (by impact):"
echo ""
for category in "${SORTED_CATEGORIES[@]}"; do
    LOC=${CATEGORY_LOC[$category]}
    FILE_COUNT=$(echo "${FILE_CATEGORIES[$category]}" | tr '|' '\n' | wc -l | tr -d ' ')
    echo "  📦 $category: $LOC LOC ($FILE_COUNT files)"
done
echo ""

# Suggest PR splits
echo "=== Recommended PR Split Strategy ==="
echo ""

PR_NUM=1
CURRENT_PR_FILES=()
CURRENT_PR_LOC=0

echo "Split into the following PRs:"
echo ""

for category in "${SORTED_CATEGORIES[@]}"; do
    CATEGORY_SIZE=${CATEGORY_LOC[$category]}

    # If adding this category would exceed 300 LOC, start a new PR
    if [ $((CURRENT_PR_LOC + CATEGORY_SIZE)) -gt 300 ] && [ $CURRENT_PR_LOC -gt 0 ]; then
        # Finalize current PR
        echo "PR $PR_NUM: ${CURRENT_PR_FILES[*]} (~$CURRENT_PR_LOC LOC)"
        echo ""

        # Generate git commands for this PR
        echo "Commands for PR $PR_NUM:"
        echo "  git checkout -b pr$PR_NUM-split $BASE_BRANCH"
        echo "  git checkout $(git rev-parse --abbrev-ref HEAD) -- ${CURRENT_PR_FILES[*]}"
        echo "  git add ${CURRENT_PR_FILES[*]}"
        echo "  git commit -m \"Part $PR_NUM: $(echo ${CURRENT_PR_FILES[@]} | sed 's/ /, /g')\""
        echo "  gh pr create --title \"Part $PR_NUM/N: [Brief description]\" --base $BASE_BRANCH"
        echo ""

        # Start new PR
        PR_NUM=$((PR_NUM + 1))
        CURRENT_PR_FILES=()
        CURRENT_PR_LOC=0
    fi

    # Add category to current PR
    CURRENT_PR_FILES+=("$category")
    CURRENT_PR_LOC=$((CURRENT_PR_LOC + CATEGORY_SIZE))
done

# Finalize last PR
if [ $CURRENT_PR_LOC -gt 0 ]; then
    echo "PR $PR_NUM: ${CURRENT_PR_FILES[*]} (~$CURRENT_PR_LOC LOC)"
    echo ""
    echo "Commands for PR $PR_NUM:"
    echo "  git checkout -b pr$PR_NUM-split $BASE_BRANCH"
    echo "  git checkout $(git rev-parse --abbrev-ref HEAD) -- ${CURRENT_PR_FILES[*]}"
    echo "  git add ${CURRENT_PR_FILES[*]}"
    echo "  git commit -m \"Part $PR_NUM: $(echo ${CURRENT_PR_FILES[@]} | sed 's/ /, /g')\""
    echo "  gh pr create --title \"Part $PR_NUM/N: [Brief description]\" --base $BASE_BRANCH"
    echo ""
fi

TOTAL_PRS=$PR_NUM

echo "=== Summary ==="
echo ""
echo "Total PRs suggested: $TOTAL_PRS"
echo "Average PR size: ~$((TOTAL_LOC / TOTAL_PRS)) LOC"
echo ""
echo "💡 Tips:"
echo "  • Create PRs in dependency order (foundations first)"
echo "  • Ensure each PR can be merged independently"
echo "  • Update PR titles with meaningful descriptions"
echo "  • Link PRs together for context"
echo ""
echo "Alternative: If files are tightly coupled, consider:"
echo "  • Splitting by feature (not file type)"
echo "  • Creating foundation PR + feature PR"
echo "  • Extracting tests into separate PR"
echo ""

exit 0
