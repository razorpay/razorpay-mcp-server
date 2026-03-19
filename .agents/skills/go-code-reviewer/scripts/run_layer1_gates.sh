#!/bin/bash
# Run all Layer 1 correctness gates with fail-fast behavior
# Exits immediately on first failure (unless in non-blocking mode)
#
# Usage: ./run_layer1_gates.sh [--non-blocking]
#
# Modes:
#   Default: BLOCKING - Exits on first failure (for CI/CD)
#   --non-blocking: ADVISORY - Reports all issues but doesn't exit (for local dev)

# Check for non-blocking flag
NON_BLOCKING=false
if [ "$1" = "--non-blocking" ] || [ "$1" = "-n" ]; then
    NON_BLOCKING=true
    shift
fi

set -e

RESULTS=""
FAILURES=0

echo "=== Layer 1: Correctness Gates ==="
echo ""
if [ "$NON_BLOCKING" = true ]; then
    echo "ℹ️  Running in NON-BLOCKING mode - reports all issues"
    echo "   (Useful for local development when environment isn't fully set up)"
else
    echo "⚡ Running with FAIL-FAST mode - stops at first error"
fi
echo ""

# Function to run a check and fail fast
run_check() {
    local name="$1"
    local cmd="$2"
    local log_file="/tmp/layer1_${name// /_}.log"

    echo "Running: $name"
    echo -n "  ⏳ "

    if eval "$cmd" > "$log_file" 2>&1; then
        echo "✅ PASS"
        RESULTS="$RESULTS\n✅ $name: PASS"
    else
        echo "❌ FAIL"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "🛑 LAYER 1 FAILED: $name"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Error output (first 30 lines):"
        echo ""
        head -30 "$log_file" | sed 's/^/  | /'
        echo ""

        local line_count=$(wc -l < "$log_file")
        if [ "$line_count" -gt 30 ]; then
            echo "  ... ($((line_count - 30)) more lines)"
            echo "  Full log: $log_file"
            echo ""
        fi

        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "❌ STOPPING AT FIRST FAILURE"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Why we stop here:"
        echo "  • No point running more checks if code doesn't build"
        echo "  • Fixes build errors first, then re-run review"
        echo "  • Saves time and avoids cascading failures"
        echo ""
        echo "How to fix:"
        echo "  1. Review the error output above"
        echo "  2. Fix the issues in your code"
        echo "  3. Run the review again"
        echo ""
        echo "Full error log available at: $log_file"
        echo ""

        exit 1  # Exit immediately on first failure
    fi
    echo ""
}

# Optional: Check for common issues first (super fast pre-flight)
echo "🔍 Pre-flight checks..."

# Check if go.mod exists
if [ ! -f go.mod ]; then
    echo "  ❌ go.mod not found - not a Go module"
    echo ""
    echo "This doesn't appear to be a Go module."
    echo "Run 'go mod init <module-name>' to initialize."
    exit 1
fi
echo "  ✅ go.mod found"

# Check Go version
if ! go version >/dev/null 2>&1; then
    echo "  ❌ Go not installed or not in PATH"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo "  ✅ Go installed: $GO_VERSION"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Running correctness checks (fail-fast mode)..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. Go Build (CRITICAL - must pass first)
run_check "go build" "go build ./..."

# 2. Go Test (if build passes)
run_check "go test" "go test ./..."

# 3. Go Vet (if tests pass)
run_check "go vet" "go vet ./..."

# 4. Race Detection (if tests exist and vet passes)
if ls *_test.go >/dev/null 2>&1 || find . -name "*_test.go" | grep -q .; then
    run_check "go test -race" "go test -race ./..."
else
    echo "No tests found, skipping race detection"
    RESULTS="$RESULTS\n⚠️  go test -race: SKIPPED (no tests)"
    echo ""
fi

# 5. golangci-lint (if available and all above passed)
if command -v golangci-lint >/dev/null 2>&1; then
    run_check "golangci-lint" "golangci-lint run ./..."
else
    echo "⚠️  golangci-lint not found, skipping"
    echo "   💡 Install with: brew install golangci-lint"
    RESULTS="$RESULTS\n⚠️  golangci-lint: SKIPPED (not installed)"
    echo ""
fi

# 6. go mod tidy check (if everything else passed)
echo "Running: go mod tidy check"
echo -n "  ⏳ "

cp go.mod go.mod.backup 2>/dev/null || true
cp go.sum go.sum.backup 2>/dev/null || true

if go mod tidy 2>/tmp/layer1_mod_tidy.log; then
    if ! diff -q go.mod go.mod.backup >/dev/null 2>&1 || ! diff -q go.sum go.sum.backup >/dev/null 2>&1; then
        echo "❌ FAIL"
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "🛑 go.mod or go.sum not tidy"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Your go.mod or go.sum files have uncommitted changes."
        echo ""
        echo "To fix:"
        echo "  1. Run: go mod tidy"
        echo "  2. Commit the changes to go.mod and go.sum"
        echo "  3. Re-run the review"
        echo ""

        mv go.mod.backup go.mod 2>/dev/null || true
        mv go.sum.backup go.sum 2>/dev/null || true

        exit 1
    else
        echo "✅ PASS"
        RESULTS="$RESULTS\n✅ go mod tidy: PASS"
        rm go.mod.backup go.sum.backup 2>/dev/null || true
    fi
else
    echo "❌ FAIL"
    echo ""
    echo "go mod tidy failed. See /tmp/layer1_mod_tidy.log"
    mv go.mod.backup go.mod 2>/dev/null || true
    mv go.sum.backup go.sum 2>/dev/null || true
    exit 1
fi
echo ""

# If we got here, everything passed!
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Layer 1: ALL CHECKS PASSED"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Summary:"
echo -e "$RESULTS"
echo ""
echo "🎉 Code is ready for Layer 2 (Scope Boundaries)"
echo ""

exit 0
