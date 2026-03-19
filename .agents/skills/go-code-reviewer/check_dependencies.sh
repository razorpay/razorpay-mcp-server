#!/bin/bash
# Check dependencies for go-code-reviewer skill

echo "=== go-code-reviewer v4.3 - Dependency Check ==="
echo ""

ALL_GOOD=true

# Required dependencies
echo "Required Dependencies:"
if command -v gh >/dev/null 2>&1; then
    echo "  ✅ gh (GitHub CLI) - $(gh --version | head -1)"
else
    echo "  ❌ gh (GitHub CLI) - MISSING"
    echo "     Install: brew install gh (macOS) or https://cli.github.com"
    ALL_GOOD=false
fi

if command -v jq >/dev/null 2>&1; then
    echo "  ✅ jq (JSON processor) - $(jq --version)"
else
    echo "  ❌ jq - MISSING"
    echo "     Install: brew install jq (macOS) or sudo apt-get install jq (Linux)"
    ALL_GOOD=false
fi

if command -v git >/dev/null 2>&1; then
    echo "  ✅ git - $(git --version)"
else
    echo "  ❌ git - MISSING"
    echo "     Install: https://git-scm.com/downloads"
    ALL_GOOD=false
fi

echo ""
echo "GitHub Authentication:"
if gh auth status >/dev/null 2>&1; then
    echo "  ✅ GitHub authenticated"
else
    echo "  ❌ Not authenticated"
    echo "     Run: gh auth login"
    ALL_GOOD=false
fi

echo ""
echo "Optional (for Layer 1 correctness gates):"
if command -v go >/dev/null 2>&1; then
    echo "  ✅ go (compiler) - $(go version)"
else
    echo "  ⚠️  go - Not installed (needed for build/test checks)"
    echo "     Install: brew install go or https://go.dev/dl/"
fi

if command -v golangci-lint >/dev/null 2>&1; then
    echo "  ✅ golangci-lint - $(golangci-lint --version | head -1)"
else
    echo "  ⚠️  golangci-lint - Not installed (optional linter)"
    echo "     Install: brew install golangci-lint"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ "$ALL_GOOD" = true ]; then
    echo "✅ All required dependencies installed!"
    echo ""
    echo "Ready to use go-code-reviewer. Try:"
    echo "  ./scripts/fetch_pr.sh <pr-number>"
    exit 0
else
    echo "❌ Missing required dependencies (see above)"
    echo ""
    echo "Install missing tools and run this script again."
    exit 1
fi
