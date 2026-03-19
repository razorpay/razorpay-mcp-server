#!/bin/bash
# Analyze PR change patterns to recommend which review agents to spawn
# Usage: ./analyze_pr_patterns.sh [base-branch]
#
# Outputs JSON with recommended agents and their prompts

set -e

BASE_BRANCH="${1:-main}"

# Get list of changed files
CHANGED_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null || echo "")

if [ -z "$CHANGED_FILES" ]; then
    echo "Error: No changes detected between $BASE_BRANCH and HEAD"
    exit 1
fi

# Get PR size metrics
TOTAL_LOC=$(git diff --shortstat "$BASE_BRANCH"...HEAD | awk '{print $4+$6}')
TOTAL_FILES=$(echo "$CHANGED_FILES" | wc -l | tr -d ' ')

echo "=== PR Change Pattern Analysis ==="
echo ""
echo "Metrics:"
echo "  Lines changed: $TOTAL_LOC"
echo "  Files changed: $TOTAL_FILES"
echo ""

# Detect change patterns
HAS_AUTH_CHANGES=$(echo "$CHANGED_FILES" | grep -i "auth\|login\|session\|token\|password\|jwt\|oauth" || echo "")
HAS_DB_CHANGES=$(echo "$CHANGED_FILES" | grep -i "repo\|database\|query\|sql\|migration\|schema" || echo "")
HAS_API_CHANGES=$(echo "$CHANGED_FILES" | grep -i "api\|handler\|server\|endpoint\|controller\|route" || echo "")
HAS_PERF_CRITICAL=$(echo "$CHANGED_FILES" | grep -i "service\|processor\|worker\|job\|batch" || echo "")
HAS_CONCURRENCY=$(git diff "$BASE_BRANCH"...HEAD | grep -i "go func\|goroutine\|channel\|sync\.\|context\." || echo "")
HAS_TEST_FILES=$(echo "$CHANGED_FILES" | grep "_test\.go$" || echo "")
HAS_TRANSACTIONS=$(git diff "$BASE_BRANCH"...HEAD | grep -i "\.Transaction(\|db\.WithTransaction\|tx\.Exec\|tx\.Query\|gorm.*Begin()" || echo "")

# Determine which agents to spawn
AGENTS_TO_SPAWN=()

# Always spawn go-reviewer for Go PRs
if echo "$CHANGED_FILES" | grep -q "\.go$"; then
    AGENTS_TO_SPAWN+=("go-reviewer")
fi

# Conditionally spawn based on patterns
if [ -n "$HAS_AUTH_CHANGES" ]; then
    AGENTS_TO_SPAWN+=("security-reviewer")
fi

if [ -n "$HAS_DB_CHANGES" ]; then
    AGENTS_TO_SPAWN+=("performance-reviewer")
    AGENTS_TO_SPAWN+=("database-reviewer")
fi

if [ -n "$HAS_API_CHANGES" ]; then
    AGENTS_TO_SPAWN+=("api-reviewer")
fi

if [ "$TOTAL_LOC" -gt 500 ] || [ "$TOTAL_FILES" -gt 10 ]; then
    AGENTS_TO_SPAWN+=("architecture-reviewer")
fi

if [ -n "$HAS_CONCURRENCY" ]; then
    AGENTS_TO_SPAWN+=("go-reviewer")  # Already added, but emphasize concurrency
fi

# Remove duplicates
AGENTS_TO_SPAWN=($(echo "${AGENTS_TO_SPAWN[@]}" | tr ' ' '\n' | sort -u))

echo "Detected patterns:"
[ -n "$HAS_AUTH_CHANGES" ] && echo "  🔐 Authentication/Security changes"
[ -n "$HAS_DB_CHANGES" ] && echo "  🗄️  Database/Repository changes"
[ -n "$HAS_TRANSACTIONS" ] && echo "  🚨 Transaction code - CRITICAL: Check transaction context usage!"
[ -n "$HAS_API_CHANGES" ] && echo "  🌐 API/Handler changes"
[ -n "$HAS_PERF_CRITICAL" ] && echo "  ⚡ Performance-critical components"
[ -n "$HAS_CONCURRENCY" ] && echo "  🔀 Concurrency patterns (goroutines/channels)"
[ -n "$HAS_TEST_FILES" ] && echo "  ✅ Test file changes"
echo ""

# Recommendation logic
if [ "$TOTAL_LOC" -lt 500 ] && [ "$TOTAL_FILES" -lt 10 ]; then
    echo "Recommendation: Manual review (PR is small enough)"
    echo ""
    echo "Agents available but not required:"
    for agent in "${AGENTS_TO_SPAWN[@]}"; do
        echo "  • backend-engineer:review:$agent"
    done
    exit 0
fi

echo "Recommendation: Spawn ${#AGENTS_TO_SPAWN[@]} specialized agent(s) in parallel"
echo ""
echo "=== Agents to Spawn ==="
echo ""

# Generate agent spawn commands
for agent in "${AGENTS_TO_SPAWN[@]}"; do
    echo "Task: backend-engineer:review:$agent"

    case "$agent" in
        go-reviewer)
            GO_FILES=$(echo "$CHANGED_FILES" | grep "\.go$" | head -20)
            echo "Prompt: \"Review Go code changes for idioms, best practices, and quality."
            if [ -n "$HAS_CONCURRENCY" ]; then
                echo "CRITICAL: This PR includes goroutines/channels - pay special attention to:"
                echo "- Race conditions and data races"
                echo "- Proper context propagation"
                echo "- Channel closure and deadlock prevention"
                echo "- Wait group usage"
                echo ""
            fi
            echo "Changed Go files:"
            echo "$GO_FILES"
            echo ""
            echo "Focus on:"
            echo "- Error handling patterns"
            echo "- Resource cleanup (defer, context)"
            echo "- Naming conventions"
            echo "- Test quality\""
            ;;

        security-reviewer)
            AUTH_FILES=$(echo "$CHANGED_FILES" | grep -i "auth\|login\|session\|token" | head -10)
            echo "Prompt: \"Review security aspects focusing on authentication and authorization."
            echo ""
            echo "Files with auth/security changes:"
            echo "$AUTH_FILES"
            echo ""
            echo "Check for:"
            echo "- Proper authentication/authorization"
            echo "- Secure credential handling"
            echo "- Input validation"
            echo "- SQL injection prevention"
            echo "- XSS prevention\""
            ;;

        performance-reviewer)
            DB_FILES=$(echo "$CHANGED_FILES" | grep -i "repo\|query\|sql" | head -10)
            echo "Prompt: \"Review database and query performance."
            echo ""

            if [ -n "$HAS_TRANSACTIONS" ]; then
                echo "⚠️  Transaction code detected - verify transaction context usage"
                echo "See references/transaction-context.md for details"
                echo ""
            fi

            echo "Files with database changes:"
            echo "$DB_FILES"
            echo ""
            echo "Check for:"
            if [ -n "$HAS_TRANSACTIONS" ]; then
                echo "- ⚠️  Transaction context usage (coordinate with database-reviewer)"
            fi
            echo "- N+1 query problems"
            echo "- Missing indexes"
            echo "- Inefficient queries"
            echo "- Proper connection pooling"
            echo "- Transaction handling\""
            ;;

        database-reviewer)
            DB_FILES=$(echo "$CHANGED_FILES" | grep -i "migration\|schema\|repo" | head -10)
            echo "Prompt: \"Review database schema changes and migrations."
            echo ""

            if [ -n "$HAS_TRANSACTIONS" ]; then
                echo "🚨 CRITICAL: Transaction code detected!"
                echo ""
                echo "MUST check transaction context usage - see references/transaction-context.md"
                echo ""
                echo "Verify ALL database operations inside Transaction() callbacks use 'tctx' not 'ctx':"
                echo "  ❌ WRONG: r.GetPayment(ctx, ...) inside Transaction() - uses outer context"
                echo "  ✅ CORRECT: r.GetPayment(tctx, ...) - uses transaction context"
                echo ""
                echo "Why this matters:"
                echo "  • Transaction isolation - operations using 'ctx' execute OUTSIDE the transaction"
                echo "  • Connection pooling - different contexts may use different DB connections"
                echo "  • Cancellation propagation - transaction rollback won't cancel 'ctx' operations"
                echo ""
                echo "Common patterns to flag:"
                echo "  • Direct calls: r.repo.GetPayment(ctx, id) inside Transaction()"
                echo "  • Helper calls: r.validateAndSave(ctx, ...) inside Transaction()"
                echo "  • Nested operations: r.updateStatus(ctx, ...) inside Transaction()"
                echo ""
            fi

            echo "Files with schema/migration/repository changes:"
            echo "$DB_FILES"
            echo ""
            echo "Check for:"
            if [ -n "$HAS_TRANSACTIONS" ]; then
                echo "- 🚨 Transaction context usage (CRITICAL - P0 issue if wrong)"
            fi
            echo "- Migration safety (reversible)"
            echo "- Index strategy"
            echo "- Data type appropriateness"
            echo "- Foreign key constraints"
            echo "- Backward compatibility\""
            ;;

        api-reviewer)
            API_FILES=$(echo "$CHANGED_FILES" | grep -i "handler\|server\|api" | head -10)
            echo "Prompt: \"Review API design and implementation."
            echo ""
            echo "Files with API changes:"
            echo "$API_FILES"
            echo ""
            echo "Check for:"
            echo "- RESTful design"
            echo "- Proper HTTP status codes"
            echo "- Error response structure"
            echo "- API versioning"
            echo "- Backward compatibility\""
            ;;

        architecture-reviewer)
            echo "Prompt: \"Review overall architecture and design decisions."
            echo ""
            echo "PR Size: $TOTAL_LOC LOC across $TOTAL_FILES files"
            echo ""
            echo "Check for:"
            echo "- Separation of concerns"
            echo "- Appropriate abstractions"
            echo "- Design pattern usage"
            echo "- Component coupling"
            echo "- Overall code organization\""
            ;;
    esac

    echo ""
    echo "---"
    echo ""
done

echo "=== Spawning Instructions ==="
echo ""
echo "Copy the Task blocks above and execute them in PARALLEL by including"
echo "all Task tool calls in a SINGLE message to maximize efficiency."
echo ""

exit 0
