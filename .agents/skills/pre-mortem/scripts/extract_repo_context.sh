#!/usr/bin/env bash
# extract_repo_context.sh — Auto-extract repo conventions for pre-mortem agents.
#
# Greps the target repo for actual metrics, logger, and tracecode patterns instead
# of asking the LLM to write JSON manually (which produces unreliable escaping).
#
# Usage:
#   ./extract_repo_context.sh <REPO_PATH> <PR_NUMBER>
#
# Outputs:
#   /tmp/premortem/$PR_NUMBER/repo_context.json

set -euo pipefail

REPO_PATH="${1:?Usage: $0 <REPO_PATH> <PR_NUMBER>}"
PR_NUMBER="${2:?Usage: $0 <REPO_PATH> <PR_NUMBER>}"
OUT_FILE="/tmp/premortem/$PR_NUMBER/repo_context.json"

# ── Validate repo path ───────────────────────────────────────────────────────
if [ ! -d "$REPO_PATH" ]; then
  echo "ERROR: REPO_PATH '$REPO_PATH' is not a directory" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT_FILE")"

# ── Metrics pattern ─────────────────────────────────────────────────────────
# Grab a real example line from the codebase (non-comment, non-test).
# Use --exclude="*_test.go" so test files are filtered at the grep level
# (grep -h suppresses filenames, making post-pipe grep -v "_test.go" ineffective).
METRICS_EXAMPLE=$(grep -rh --include="*.go" --exclude="*_test.go" \
  -e "metrics\." -e "prometheus\." \
  "$REPO_PATH" \
  --exclude-dir=vendor --exclude-dir=.git --exclude-dir=mocks \
  2>/dev/null \
  | grep -v "^\s*//" \
  | grep -m 1 "Count\|Increment\|Gauge\|Histogram\|Inc(" \
  | sed 's/^[[:space:]]*//' \
  || echo "")

# ── Trace codes file ─────────────────────────────────────────────────────────
TRACE_FILE=$(find "$REPO_PATH" -name "*.go" \
  \( -ipath "*/tracecode*" -o -ipath "*/trace_code*" -o -ipath "*/errcode*" \) \
  -not -path "*/vendor/*" \
  2>/dev/null | head -1 || echo "")

# Grab one example constant from the trace codes file.
# Use grep -E (ERE) with + instead of BRE \+ which is a GNU extension that
# fails on macOS BSD grep.
TRACE_EXAMPLE=""
if [ -n "$TRACE_FILE" ]; then
  TRACE_EXAMPLE=$(grep -Eh "^\s+[A-Z_][A-Z_0-9]+ +=" "$TRACE_FILE" 2>/dev/null \
    | head -1 | sed 's/^[[:space:]]*//' || echo "")
fi

# ── Logger pattern ───────────────────────────────────────────────────────────
# Same --exclude="*_test.go" fix as METRICS_EXAMPLE above.
LOGGER_EXAMPLE=$(grep -rh --include="*.go" --exclude="*_test.go" \
  -e "logger\.Error\b" -e "log\.Error\b" -e "logger\.Errorw\b" -e "logger\.Errorf\b" \
  "$REPO_PATH" \
  --exclude-dir=vendor --exclude-dir=.git --exclude-dir=mocks \
  2>/dev/null \
  | grep -v "^\s*//" \
  | grep -m 1 "(" \
  | sed 's/^[[:space:]]*//' \
  || echo "")

# ── Repo skill dir ───────────────────────────────────────────────────────────
REPO_SKILL_DIR=$(find "$REPO_PATH" -type d \
  \( -path "*/.claude/skills/*-skill" -o -path "*/.agents/skills/*-skill" \) \
  2>/dev/null | head -1 || echo "")

# ── CLAUDE.md path ───────────────────────────────────────────────────────────
# Include path if CLAUDE.md exists so agents can Read it for repo conventions.
CLAUDE_MD_PATH=""
if [ -f "${REPO_PATH}/CLAUDE.md" ]; then
  CLAUDE_MD_PATH="${REPO_PATH}/CLAUDE.md"
fi

# ── Write JSON via Python (safe serialization — no manual escaping) ──────────
python3 - "$METRICS_EXAMPLE" "$TRACE_FILE" "$TRACE_EXAMPLE" \
           "$LOGGER_EXAMPLE" "$REPO_SKILL_DIR" "$CLAUDE_MD_PATH" "$OUT_FILE" <<'PYEOF'
import json, sys

(metrics_pattern, trace_codes_location, trace_code_example,
 logger_pattern, repo_skill_dir, claude_md_path, out_file) = sys.argv[1:]

context = {
    "metrics_pattern":      metrics_pattern,
    "trace_codes_location": trace_codes_location,
    "trace_code_example":   trace_code_example,
    "logger_pattern":       logger_pattern,
    "repo_skill_dir":       repo_skill_dir,
    "claude_md_path":       claude_md_path,
}

with open(out_file, "w") as f:
    json.dump(context, f, indent=2)
print(json.dumps(context, indent=2))
PYEOF

echo ""
echo "✓ Repo context extracted → $OUT_FILE"
