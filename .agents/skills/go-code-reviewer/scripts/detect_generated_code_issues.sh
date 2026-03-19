#!/bin/bash
# Detect common issues with generated code (protobufs, mocks, etc.)
# Usage: ./detect_generated_code_issues.sh

set -e

ISSUES_FOUND=0

echo "=== Detecting Generated Code Issues ==="
echo ""

# Function to report an issue
report_issue() {
    local severity="$1"
    local message="$2"
    local suggestion="$3"

    echo "$severity $message"
    echo "   💡 $suggestion"
    echo ""
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
}

# 1. Check for .proto files without generated .pb.go files
echo "Checking protobuf generation..."
if find . -name "*.proto" -type f | grep -q .; then
    PROTO_COUNT=$(find . -name "*.proto" -type f | wc -l | tr -d ' ')
    PB_GO_COUNT=$(find . -name "*.pb.go" -type f 2>/dev/null | wc -l | tr -d ' ')

    echo "  Found $PROTO_COUNT .proto files and $PB_GO_COUNT .pb.go files"

    if [ "$PROTO_COUNT" -gt 0 ] && [ "$PB_GO_COUNT" -eq 0 ]; then
        report_issue "⚠️  WARNING:" \
            "Found .proto files but NO generated .pb.go files" \
            "Run 'make proto-gen' or 'buf generate' or check build instructions"
    elif [ "$PB_GO_COUNT" -gt 0 ]; then
        echo "  ✅ Protobuf files appear to be generated"
    fi
else
    echo "  ℹ️  No .proto files found"
fi
echo ""

# 2. Check for go:generate directives without generated files
echo "Checking go:generate directives..."
if grep -r "//go:generate" . --include="*.go" 2>/dev/null | grep -q .; then
    GENERATE_COUNT=$(grep -r "//go:generate" . --include="*.go" | wc -l | tr -d ' ')
    echo "  Found $GENERATE_COUNT //go:generate directives"

    # Check for mockgen specifically
    if grep -r "//go:generate.*mockgen" . --include="*.go" 2>/dev/null | grep -q .; then
        MOCK_DIRS=$(grep -r "//go:generate.*mockgen" . --include="*.go" | sed 's/:.*//g' | xargs dirname | sort -u)

        MISSING_MOCKS=0
        for dir in $MOCK_DIRS; do
            # Check if corresponding mock files exist
            if [ ! -d "$dir/mocks" ] && [ ! -d "$dir/../mocks" ]; then
                MISSING_MOCKS=$((MISSING_MOCKS + 1))
            fi
        done

        if [ $MISSING_MOCKS -gt 0 ]; then
            report_issue "⚠️  WARNING:" \
                "Found mockgen directives but mock files may be missing" \
                "Run 'go generate ./...' to generate mocks"
        else
            echo "  ✅ Mock files appear to be present"
        fi
    fi
else
    echo "  ℹ️  No //go:generate directives found"
fi
echo ""

# 3. Check go.mod for internal package references that might not exist
echo "Checking for missing internal packages..."
if [ -f go.mod ]; then
    MODULE_NAME=$(grep "^module " go.mod | awk '{print $2}')

    # Find imports of internal packages from this module
    INTERNAL_IMPORTS=$(grep -rh "^\s*\"$MODULE_NAME/" . --include="*.go" 2>/dev/null | \
        sed 's/^\s*"//g' | sed 's/".*$//g' | sort -u || true)

    if [ -n "$INTERNAL_IMPORTS" ]; then
        MISSING_COUNT=0
        while IFS= read -r import; do
            # Convert import path to file path
            FILE_PATH=$(echo "$import" | sed "s|$MODULE_NAME/||g")

            # Check if the directory exists
            if [ ! -d "$FILE_PATH" ]; then
                if [ $MISSING_COUNT -eq 0 ]; then
                    echo "  ⚠️  Missing internal packages detected:"
                fi
                echo "    - $import"
                MISSING_COUNT=$((MISSING_COUNT + 1))
            fi
        done <<< "$INTERNAL_IMPORTS"

        if [ $MISSING_COUNT -gt 0 ]; then
            report_issue "⚠️  WARNING:" \
                "Found imports to $MISSING_COUNT internal packages that don't exist" \
                "These may be generated packages. Check if you need to run codegen commands."
        else
            echo "  ✅ All internal package imports appear valid"
        fi
    fi
fi
echo ""

# 4. Check for common build system files and suggest commands
echo "Detecting build system..."
BUILD_COMMANDS=""

if [ -f Makefile ]; then
    echo "  📋 Found Makefile"

    # Parse Makefile to extract relevant targets with their commands
    echo ""
    echo "  📝 Parsing Makefile for code generation targets..."

    # Extract all target names and their line numbers
    TARGETS=$(grep -n "^[a-zA-Z0-9_-]*:.*" Makefile | grep -E "(proto|gen|mock|buf)" | head -10)

    if [ -n "$TARGETS" ]; then
        while IFS= read -r target_line; do
            LINE_NUM=$(echo "$target_line" | cut -d: -f1)
            TARGET_NAME=$(echo "$target_line" | cut -d: -f2 | sed 's/:.*//g' | tr -d ' \t')

            # Extract the command(s) for this target (next non-empty, non-comment line)
            COMMAND=$(awk -v line="$LINE_NUM" '
                NR > line && NF > 0 && !/^#/ && /^\t/ {
                    gsub(/^\t/, "");
                    print;
                    exit;
                }
            ' Makefile)

            if [ -n "$COMMAND" ]; then
                # Truncate long commands
                SHORT_CMD=$(echo "$COMMAND" | head -c 60)
                if [ ${#COMMAND} -gt 60 ]; then
                    SHORT_CMD="${SHORT_CMD}..."
                fi

                BUILD_COMMANDS="$BUILD_COMMANDS\n  - make $TARGET_NAME"
                BUILD_COMMANDS="$BUILD_COMMANDS\n    └─ Line $LINE_NUM: $SHORT_CMD"
            fi
        done <<< "$TARGETS"
    fi

    # Also check for common patterns not yet captured
    if grep -q "^proto-gen:" Makefile && [ -z "$(echo -e "$BUILD_COMMANDS" | grep 'proto-gen')" ]; then
        BUILD_COMMANDS="$BUILD_COMMANDS\n  - make proto-gen   # Generate protobufs"
    fi

    if grep -q "^generate:" Makefile && [ -z "$(echo -e "$BUILD_COMMANDS" | grep 'generate')" ]; then
        BUILD_COMMANDS="$BUILD_COMMANDS\n  - make generate    # Generate code"
    fi

    if grep -q "^mocks:" Makefile && [ -z "$(echo -e "$BUILD_COMMANDS" | grep 'mocks')" ]; then
        BUILD_COMMANDS="$BUILD_COMMANDS\n  - make mocks       # Generate mocks"
    fi

    if grep -q "^mock-gen:" Makefile && [ -z "$(echo -e "$BUILD_COMMANDS" | grep 'mock-gen')" ]; then
        BUILD_COMMANDS="$BUILD_COMMANDS\n  - make mock-gen    # Generate mocks"
    fi
fi

if [ -f buf.yaml ] || [ -f buf.gen.yaml ]; then
    echo "  📋 Found buf configuration"
    BUILD_COMMANDS="$BUILD_COMMANDS\n  - buf generate     # Generate protobufs with buf"
fi

if [ -f .tool-versions ] || [ -f .go-version ]; then
    echo "  📋 Found version management files"
fi

# Check for scripts directory with generation scripts
if [ -d scripts ]; then
    GEN_SCRIPTS=$(find scripts -name "*gen*.sh" -o -name "*generate*.sh" | head -5)
    if [ -n "$GEN_SCRIPTS" ]; then
        echo "  📋 Found generation scripts in scripts/"
        while IFS= read -r script; do
            if [ -n "$script" ]; then
                BUILD_COMMANDS="$BUILD_COMMANDS\n  - ./$script   # Run generation script"
            fi
        done <<< "$GEN_SCRIPTS"
    fi
fi

if [ -n "$BUILD_COMMANDS" ]; then
    echo ""
    echo "  💡 Suggested build commands to run BEFORE testing:"
    echo -e "$BUILD_COMMANDS"
    echo ""
    echo "  📖 Full Makefile targets reference:"
    echo "     View Makefile to see all available targets and their documentation"
fi
echo ""

# 5. Try to detect if go.mod is out of sync
echo "Checking go.mod sync..."
if go list -m all >/dev/null 2>&1; then
    echo "  ✅ go.mod appears valid"
else
    report_issue "⚠️  WARNING:" \
        "go.mod may be out of sync" \
        "Run 'go mod tidy' to sync dependencies"
fi
echo ""

# Summary
echo "=== Summary ==="
if [ $ISSUES_FOUND -eq 0 ]; then
    echo "✅ No generated code issues detected"
    echo ""
    exit 0
else
    echo "⚠️  Found $ISSUES_FOUND potential issue(s) with generated code"
    echo ""
    echo "These issues may cause build failures in Layer 1."
    echo "Consider fixing them before running the full review."
    echo ""
    exit 2  # Exit code 2 = warnings (not critical failure)
fi
