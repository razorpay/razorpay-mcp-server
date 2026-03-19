# Go Code Reviewer v4.3

> Five-layer Go code review framework aligned with Google Engineering Practices, featuring CRITICAL transaction context detection and pragmatic approval decision logic.

## 🚀 Quick Start

### Prerequisites Check

```bash
# Run dependency check
chmod +x check_dependencies.sh
./check_dependencies.sh
```

### Review a PR in 3 Commands

```bash
# 1. Fetch the PR
./scripts/fetch_pr.sh 123

# 2. Check PR size (with GitHub API verification)
./scripts/check_pr_size.sh 123 master

# 3. Run correctness gates
./scripts/run_layer1_gates.sh
```

### Post Review with Inline Comments

```bash
./scripts/post_review.sh 123 review.md approve inline-comments.json
```

## ✨ What's New in v4.3

### 🎯 Google Approval Decision Logic (NEW!)

**"Favor approving a CL once it improves code health"**

- **Pragmatic approach**: Approves good code with minor imperfections
- **P2 truly optional**: Non-actionable suggestions, not "should fix"
- **Scope as feedback**: PR size issues don't block well-implemented code
- **Context recognition**: Values author documentation (Layer 0)

### ⚡ Simplicity & Performance

- **Scripts**: 11 → 9 (consolidated duplicates)
- **Review time**: ~8 minutes (70% faster than manual)
- **Review length**: 260-340 lines (concise, focused)
- **Inline comments**: 3-5 per review (mandatory, balanced)

### 📚 Enhanced Documentation

- **New reference**: `review-quality-standards.md` (350+ lines)
- **Total guides**: 13 (5,574+ lines)
- **Quality standards**: Google, Uber, Razorpay aligned

## 📋 Five-Layer Framework

### Layer 0: Intent (5-15 min)
Understand WHAT and WHY before reviewing HOW

### Layer 1: Correctness (10-20 min) - FAIL FAST
- ✅ Build validation
- ✅ Test execution
- 🚨 **Transaction context verification** (CRITICAL!)
- ✅ Lint checks

### Layer 2: Scope (5-10 min)
- GitHub API verification (source of truth)
- PR size analysis
- Split recommendations if needed

### Layer 3: Quality (20-40 min)
- Pattern-based agent spawning
- Specialized review agents
- Severity classification: 🚨⚠️💡

### Layer 4: Post-Review (5-10 min)
- GitHub integration with inline comments
- Review history tracking
- Metrics collection

## 🛠️ Scripts (9 Total)

### Core Review
1. `analyze_pr_patterns.sh` - Pattern detection + transaction checks
2. `run_layer1_gates.sh` - Fail-fast correctness gates
3. `check_pr_size.sh` - Scope analysis with GitHub API

### GitHub Integration
4. `fetch_pr.sh` - PR retrieval
5. `post_review.sh` - Review posting with inline comments

### Helpers
6. `generate_inline_comments_template.sh` - Comment skeleton
7. `suggest_pr_split.sh` - Smart PR splitting
8. `track_review_history.sh` - Metrics tracking
9. `detect_generated_code_issues.sh` - Generated code detection

## 📚 References (13 Guides, 5,574+ Lines)

### Original (3 guides)
- `google-go-patterns.md` - Google Go style guide
- `uber-go-patterns.md` - Uber Go style guide
- `layer3-checklist.md` - Quality review checklist

### Razorpay (7 guides, 4,577 lines)
- 🚨 `razorpay-transaction-context.md` - **CRITICAL** DB patterns (386 lines)
- `razorpay-concurrency-patterns.md` - Goroutines, channels (825 lines)
- `razorpay-error-handling.md` - Error wrapping (654 lines)
- `razorpay-performance-optimization.md` - Memory, I/O (824 lines)
- `razorpay-testing-best-practices.md` - Table-driven tests (721 lines)
- `razorpay-idiomatic-go.md` - Packages, structs (860 lines)
- `QUICK_REFERENCE.md` - 60-second checklist (307 lines)

### v4.2+ (3 guides)
- `solid-principles-go.md` - SOLID in Go (324 lines)
- `security-assessment-checklist.md` - Security risks (323 lines)
- 🎯 `review-quality-standards.md` - Google Engineering Practices (350+ lines)

## 🚨 Critical Feature: Transaction Context Detection

Automatically detects **critical database transaction bugs**:

```go
// ❌ CRITICAL BUG - Detected automatically!
return r.db.Transaction(ctx, func(tctx context.Context) error {
    payment, err := r.GetPayment(ctx, id)  // Using ctx instead of tctx!
})

// ✅ CORRECT
return r.db.Transaction(ctx, func(tctx context.Context) error {
    payment, err := r.GetPayment(tctx, id)  // Using tctx!
})
```

**Why this matters**:
- Transaction isolation violations
- Different DB connections possible
- Rollback won't cancel ctx operations
- **Potential data corruption**

## 💡 Usage Examples

### Example 1: Full Review Flow

```bash
# Fetch PR with automatic branch setup
./scripts/fetch_pr.sh 123
# ✅ PR #123 ready for review
# Details: 4 files, 286 lines

# Check size (GitHub API as source of truth)
./scripts/check_pr_size.sh 123 master
# Status: ⚠️ WARNING - 286 lines (target: <200)
# ✅ VERIFIED: GitHub API and git diff match

# Run Layer 1 gates
./scripts/run_layer1_gates.sh
# ✅ Build: PASSED
# ✅ Tests: PASSED
# ✅ Lint: PASSED

# Analyze patterns
./scripts/analyze_pr_patterns.sh master
# 🗄️ Database/Repository changes detected
# 🚨 Transaction code - CRITICAL: Check transaction context usage!
# Recommendation: Spawn database-reviewer agent

# Post review with inline comments
./scripts/post_review.sh 123 review.md request-changes inline.json
# ✅ Review posted with 5 inline comments
```

### Example 2: Real Review Results (PR #110)

**Found**: Critical P0 unique index bug (blocks soft delete)
**Time**: 8 minutes
**Review**: 260 lines (concise)
**Inline comments**: 5 (3 posted successfully)
**Verdict**: REQUEST CHANGES with clear fix guidance (<5 min fix)

### Example 3: 60-Second Quick Review

```bash
cat QUICK_REFERENCE.md
```

**Quick scan pattern**:
1. Search `Transaction(` → verify context (30s)
2. Search `go func(` → check leaks (15s)
3. Search `make(chan` → check buffering (10s)
4. Search `defer` → verify cleanup (5s)

## 🎯 Severity Classification

- 🚨 **Critical (P0)** - Must fix before merge
  - Transaction context bugs
  - Security vulnerabilities
  - Data corruption risks
  - Failing tests/build

- ⚠️ **Important (P1)** - Should fix (non-blocking for good code)
  - Missing error wrapping
  - Race conditions
  - N+1 queries
  - Scope issues (PR too large)

- 💡 **Optional (P2)** - Nice to have (truly non-actionable)
  - Variable naming
  - Code organization
  - Speculative improvements

## ⏱️ Time Savings

**Real-world validation**:
- **Manual review**: 25-30 minutes per PR
- **With v4.3**: ~8 minutes per PR
- **Time saved**: 70% (17-22 minutes)
- **Focus shift**: Finding issues → Validating findings

## ✅ Dependencies

### Required
- **gh** (GitHub CLI) - `brew install gh`
- **jq** (JSON processor) - `brew install jq`
- **git** - Usually pre-installed

### For Go Projects (Layer 1 gates)
- **go** (1.18+) - `brew install go`
- **golangci-lint** (optional) - `brew install golangci-lint`

### Setup
```bash
# Install dependencies (macOS)
brew install gh jq go golangci-lint

# Authenticate with GitHub
gh auth login

# Verify setup
./check_dependencies.sh
```

## 🤖 Agent Integration

Works seamlessly with specialized agents:

```bash
# Spawn multiple agents in parallel (single message)
Task: backend-engineer:review:database-reviewer
Task: backend-engineer:review:security-reviewer
Task: backend-engineer:review:go-reviewer
```

## 📊 Output Guardrails

### Professional Reviews
- **Length**: <350 lines (typical), <450 (large PR)
- **Emojis**: 10-15 max (status only: ✅❌⚠️💡🚨)
- **Tone**: Professional, no casual language
- **Inline comments**: 3-5 per review (mandatory)

### Quality Standards
Based on research from:
- ✅ Google Go Style Guide
- ✅ Uber Go Style Guide
- ✅ Razorpay Best Practices
- ✅ ISO/IEC 25010 Software Quality
- ✅ **Google Engineering Practices (Code Review Guide)**

## 🏆 Version History

- **v4.3** (2026-02-05) - Google approval logic, simplicity cleanup
- **v4.2** (2026-02-04) - Enterprise review structure, SOLID assessment
- **v4.1** (2026-02-04) - Guardrails, mandatory inline comments
- **v4.0** (2026-02-04) - Razorpay integration, transaction detection
- **v3.0** - Phase 2 automation, agent spawning
- **v2.0** - GitHub integration, review history
- **v1.0** - Initial five-layer framework

## 🔧 Configuration

Scripts use environment variables (optional):

```bash
export REVIEW_BASE_BRANCH="main"        # Default base branch
export REVIEW_MIN_COVERAGE="80"         # Minimum test coverage
export REVIEW_MAX_PR_SIZE="500"         # Max LOC before split
```

## 🐛 Troubleshooting

### Dependency issues
```bash
./check_dependencies.sh
# Shows what's missing and how to install
```

### GitHub posting fails
```bash
gh auth status              # Check authentication
gh auth login               # Re-authenticate if needed
```

### Inline comments not appearing
```bash
# Check PR HEAD commit SHA
gh pr view <pr> --json headRefOid -q .headRefOid

# Verify commit matches the one being reviewed
```

## 📈 CI/CD Integration

### GitHub Actions
```yaml
# .github/workflows/code-review.yml
- name: Run Layer 1 Gates
  run: ./scripts/run_layer1_gates.sh

- name: Check PR Size
  run: ./scripts/check_pr_size.sh ${{ github.event.number }} main
```

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
./scripts/run_layer1_gates.sh || exit 1
```

## 🌟 Best Practices

1. **Check dependencies first**: Run `./check_dependencies.sh`
2. **Run Layer 1 gates early**: Fail fast on critical issues
3. **Verify transaction patterns**: Data corruption is critical
4. **Use GitHub API metrics**: Source of truth for PR size
5. **Create inline comments**: Direct feedback on code lines
6. **Track review history**: Continuous improvement

## 📖 Documentation

- **SKILL.md** - Complete skill documentation (750 lines)
- **README.md** - This file (quick start guide)
- **QUICK_REFERENCE.md** - 60-second review checklist
- **RELEASE-NOTES-v4.*.md** - Version-specific changes
- **check_dependencies.sh** - Dependency verification script

## 💬 Support

- All scripts have `--help` flag
- Reference docs have detailed examples
- Pattern detection provides guidance
- Review history tracks metrics

## 📄 License

Internal use only. Reference documentation attributed to respective sources.

---

**Ready to review Go code with Google standards?**

```bash
./check_dependencies.sh && ./scripts/fetch_pr.sh <pr-number>
```

🎯 **Time saved**: 70% | 📝 **Review quality**: Research-backed | 🚨 **Critical bugs**: Auto-detected
