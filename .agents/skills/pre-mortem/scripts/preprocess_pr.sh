#!/usr/bin/env bash
# preprocess_pr.sh — Fetch, filter, and cache a PR diff by check category.
#
# Usage:
#   ./preprocess_pr.sh <PR_NUMBER> [CACHE_DIR]
#
# Outputs:
#   $CACHE_DIR/diff.txt              — full raw diff (cached by commit SHA)
#   $CACHE_DIR/changed_files.txt     — newline-separated changed file paths
#   $CACHE_DIR/infra_diff.txt        — diff sections relevant to infrastructure checks
#   $CACHE_DIR/services_diff.txt     — diff sections relevant to service integration checks
#   $CACHE_DIR/domain_quality_diff.txt — diff sections relevant to domain+quality checks
#   $CACHE_DIR/observability_diff.txt  — all non-test Go file diffs (observability always runs)
#   $CACHE_DIR/manifest.json         — which categories have content + line counts
#
# Token savings vs inline approach:
#   - Diff fetched once and cached; re-runs skip gh pr diff entirely
#   - Agents receive only their category's filtered slice (~80-90% reduction per agent)
#   - manifest.json tells the main agent which sub-agents to skip

set -euo pipefail

PR_NUMBER="${1:?Usage: $0 <PR_NUMBER> [CACHE_DIR]}"
CACHE_DIR="${2:-/tmp/premortem/$PR_NUMBER}"

mkdir -p "$CACHE_DIR"

# ─── Cache validity check ──────────────────────────────────────────────────────
# Use the PR's HEAD commit SHA as the cache key. If the PR is updated,
# the SHA changes and we re-fetch. This makes re-runs on the same PR instant.

CURRENT_SHA=$(gh pr view "$PR_NUMBER" --json headRefOid --jq .headRefOid 2>/dev/null || echo "unknown")
SHA_FILE="$CACHE_DIR/.sha"

if [ -f "$SHA_FILE" ] && [ -f "$CACHE_DIR/diff.txt" ] && [ "$(cat "$SHA_FILE")" = "$CURRENT_SHA" ]; then
  echo "✓ Cache hit for PR #$PR_NUMBER @ $CURRENT_SHA — skipping re-fetch"
  cat "$CACHE_DIR/manifest.json"
  exit 0
fi

echo "→ Fetching PR #$PR_NUMBER (SHA: $CURRENT_SHA)..."

# ─── Fetch diff and file list ──────────────────────────────────────────────────
gh pr diff "$PR_NUMBER" > "$CACHE_DIR/diff.txt"
gh pr diff "$PR_NUMBER" --name-only > "$CACHE_DIR/changed_files.txt"

# ─── Filter diff into category slices ─────────────────────────────────────────
python3 - "$CACHE_DIR" <<'PYEOF'
import sys, re, json
from pathlib import Path

out = Path(sys.argv[1])
diff_text = (out / "diff.txt").read_text(errors="replace")

# Split full diff into per-file sections
# Each section starts with "diff --git a/<file> b/<file>"
sections = re.split(r'(?=^diff --git )', diff_text, flags=re.MULTILINE)
sections = [s for s in sections if s.strip()]

# ── Category detection patterns ──────────────────────────────────────────────
# Applied to BOTH file path AND diff content so files in non-standard locations
# are still routed correctly (e.g. splitz.GetVariant used in a domain file).

INFRA_PATH = re.compile(r"""
    internal/[^/]+/repo\.go          # generic DB repos
  | pkg/db/                          # DB clients
  | pkg/dynamodb/                    # DynamoDB
  | internal/kafka/                  # Kafka consumers
  | worker/kafka/                    # Kafka workers
  | internal/cache/                  # Redis cache
  | pkg/queue/                       # Redis/SQS queue clients
  | worker/dispatcher/               # SQS dispatchers
  | internal/events?/                # eventing
  | internal/event_                  # event handlers
  | pkg/httpclient/                  # HTTP clients (resilience)
  | configs/.*\.toml$                # TOML configs
""", re.VERBOSE)

INFRA_CONTENT = re.compile(r"""
    \bgorm\.DB\b | \.Begin\(\) | \.Rollback\(\) | \.Commit\(\)  # transactions
  | dynamodb\. | kafka\. | redis\. | sqs\.                       # infra clients
  | circuit[Bb]reaker | WithTimeout | context\.WithDeadline      # resilience
""", re.VERBOSE)

SERVICES_PATH = re.compile(r"""
    internal/mozart/  | internal/splitz/  | pkg/splitz/
  | internal/stork/   | pkg/stork/
  | internal/passport/| pkg/passport/
  | internal/account[_s]/ | pkg/account[_s]/
  | internal/router/  | pkg/router/
  | internal/pgrouter/| pkg/pgrouter/
""", re.VERBOSE)

SERVICES_CONTENT = re.compile(r"""
    splitz\.GetVariant | splitz\.Client                           # Splitz
  | stork\.NewClient | SendSMS | SendEmail | SendWhatsApp         # Stork
  | passport\.InitHandler | FromToken | X-Passport-JWT           # Passport
  | accountService\. | \.GetByID\( | \.Write\(\)\.Save           # ASV
  | router\.NewClient | GetTerminals | NewFetchTerminals          # Router SDK
  | MutexClient | PaymentCallback                                 # PG-Router
""", re.VERBOSE)

DOMAIN = re.compile(r"""
    internal/[^/]+/                   # any internal domain package
  | \w+_test\.go$                     # unit test files
  | ^slit/                            # integration test files
""", re.VERBOSE)

# Observability runs on all non-test Go source files
OBS = re.compile(r'\.go$')
TEST = re.compile(r'_test\.go$')

# ── Categorise each section ───────────────────────────────────────────────────
# A section can appear in multiple buckets (non-exclusive).
# Content-based matching catches service usage in non-standard file locations.
buckets = {"infra": [], "services": [], "domain_quality": [], "observability": []}

for section in sections:
    m = re.match(r'diff --git a/(\S+)', section)
    fname = m.group(1) if m else ""

    # Match on file path OR diff content to avoid false negatives from
    # repos with non-standard directory layouts.
    if INFRA_PATH.search(fname) or INFRA_CONTENT.search(section):
        buckets["infra"].append(section)
    if SERVICES_PATH.search(fname) or SERVICES_CONTENT.search(section):
        buckets["services"].append(section)
    # Domain is a catchall — any file not already in infra/services lands here,
    # ensuring cross-category interactions are visible to at least one agent.
    if DOMAIN.search(fname) or (fname.endswith('.go') and not TEST.search(fname)):
        buckets["domain_quality"].append(section)
    if OBS.search(fname) and not TEST.search(fname):
        buckets["observability"].append(section)

# ── Write filtered files ──────────────────────────────────────────────────────
manifest = {}
for name, parts in buckets.items():
    path = out / f"{name}_diff.txt"
    content = "\n".join(parts)
    path.write_text(content)
    manifest[name] = {
        "has_content": bool(parts),
        "files":       len(parts),
        "lines":       content.count("\n"),
    }

(out / "manifest.json").write_text(json.dumps(manifest, indent=2))
print(json.dumps(manifest, indent=2))
PYEOF

# ── Save SHA after successful processing ──────────────────────────────────────
echo "$CURRENT_SHA" > "$SHA_FILE"

echo ""
echo "✓ Preprocessed PR #$PR_NUMBER → $CACHE_DIR"
echo "  infra:             $(wc -l < "$CACHE_DIR/infra_diff.txt") lines"
echo "  services:          $(wc -l < "$CACHE_DIR/services_diff.txt") lines"
echo "  domain_quality:    $(wc -l < "$CACHE_DIR/domain_quality_diff.txt") lines"
echo "  observability:     $(wc -l < "$CACHE_DIR/observability_diff.txt") lines"
