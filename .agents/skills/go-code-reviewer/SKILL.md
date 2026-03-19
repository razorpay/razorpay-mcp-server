---
name: go-code-reviewer
description: "Five-layer Go code review framework (Intent → Correctness → Scope → Quality → Post) with CRITICAL transaction context detection and mandatory inline comments. Use when reviewing Go code changes, PRs, or commits. Triggers on: 'review this PR', 'check my Go code', 'review my changes', '/go-code-reviewer', '/review-go', or when user asks for Go code quality assessment. Enforces research-backed quality gates from Google, Uber, and Razorpay standards with strict output guardrails (concise reviews under 300 lines, professional tone, 3-5 mandatory inline comments per review). Includes automated pattern detection for auth, database, API, and concurrency changes. NEW in v4.2: Executive summary with quality ratings, SOLID principles assessment, security risk levels, positive observations section, acceptance criteria, and structured next steps. v4.1: Mandatory inline comments and guardrails. v4.0: Transaction context bug detection (CRITICAL), comprehensive Razorpay reference guides."
---

# Go Code Reviewer v4.2

Comprehensive five-layer Go code review framework with automated pattern detection, GitHub integration, and research-backed quality gates. Combines automation scripts with deep reference documentation from Google, Uber, and Razorpay best practices.

## Version 4.2 Highlights

**NEW in v4.2** - Enterprise-Grade Review Structure:
- ⭐ **Executive Summary**: Instant clarity with quality ratings (Excellent → Poor), production readiness, security risk level, and issue counts
- ⭐ **SOLID Principles Assessment**: Systematic architectural quality evaluation with checklist
- ⭐ **Security Risk Rating**: Explicit security assessment (None/Low/Medium/High/Critical) with vulnerability analysis
- ⭐ **Positive Observations**: Dedicated section recognizing excellent patterns (balanced feedback)
- ⭐ **Acceptance Criteria**: Clear merge requirements (Must/Should/Nice to Have)
- ⭐ **Structured Next Steps**: Prioritized recommendations (Immediate/Short-term/Long-term/Tooling/Documentation)

**Previous Versions**:
- **v4.1**: Mandatory inline comments (3-5 per review) + restored output guardrails
- **v4.0**: Transaction context detection (CRITICAL) + 7 Razorpay reference guides (4,577 lines covering concurrency, error handling, performance, testing, idiomatic Go)

## Triggers

This skill activates when users:
- Say "review this PR" or "check my Go code"
- Use `/go-code-reviewer` or `/review-go`
- Ask for Go code quality assessment
- Request PR review or code review
- Mention "check my changes" for Go code

## Pre-Review Checklist (MANDATORY)

**Before starting Layer 0, verify**:

- [ ] PR fetched correctly: `gh pr view <pr>` shows expected PR
- [ ] Working directory clean: `git status` shows no uncommitted changes
- [ ] Correct branch checked out: `git branch --show-current` matches PR branch

**If any check fails**:
1. Stop review
2. Clean working directory: `git stash` or `git reset --hard`
3. Re-fetch PR: `./scripts/fetch_pr.sh <pr>`
4. Re-verify checklist

**Note**: PR metrics verification happens automatically in Layer 2 via `check_pr_size.sh`, which uses GitHub API as source of truth and warns on mismatches.

---

## Five-Layer Review Framework

### Layer 0: Intent Verification (5-15 min)
**Purpose**: Understand WHAT and WHY before reviewing HOW
- Read PR description carefully - check if author explained the change
- Confirm stated intent aligns with implementation
- **Important**: If PR description provides context, acknowledge it before questioning
  - ✅ "Author documented X in description. Confirmed."
  - ❌ "Why is X included?" (when already explained)
- Only flag missing context if truly absent
- **Output**: Clear understanding of change purpose with recognition of provided context

### Layer 1: Correctness Gates (10-20 min) - FAIL FAST
**Purpose**: Catch critical bugs that could break production
- ✅ Build validation (compilation check)
- ✅ Test execution (unit + integration)
- 🚨 **NEW**: Transaction context verification (CRITICAL)
- ✅ Basic lint checks (gofmt, go vet)
- **Action**: STOP if any gate fails - fix before proceeding

### Layer 2: Scope Analysis (5-10 min)
**Purpose**: Assess PR size and focus, provide split recommendations if needed

**CRITICAL - VERIFICATION REQUIRED**:
1. **Run scope analysis** with automatic GitHub API verification
   ```bash
   ./scripts/check_pr_size.sh <pr-number>
   ```
2. **Script automatically**:
   - Uses GitHub API as source of truth
   - Verifies against local git diff
   - Warns if mismatch detected
3. **If mismatch detected**: Use GitHub API values shown in output, NOT git diff

**Scope Analysis**:
- PR size metrics (LOC changed, files touched)
- Scope appropriateness (single concern vs multiple)
- Generated code detection
- **Recommendation**: Split if >500 LOC or >10 files

**Important**: Scope issues (multiple concerns, large PRs) are **architectural feedback**, not merge blockers. They indicate process improvements but should not block well-implemented code.

**Sanity Checks** (MANDATORY):
- Does file count match PR title/description?
- Are cross-domain changes explained in PR description?
- Does LOC seem reasonable for stated changes?
- **Red flags** (require extra verification):
  - File count >3x expected from PR title
  - Cross-domain files in single-domain PR without explanation
  - LOC >500 for "simple" changes

### Layer 3: Quality Review (20-40 min)
**Purpose**: Deep code quality assessment with pattern matching
- **Automated pattern detection** spawns specialized agents:
  - `go-reviewer`: Go idioms, error handling, testing
  - `security-reviewer`: Auth, credentials, input validation
  - `performance-reviewer`: N+1 queries, indexing, caching
  - `database-reviewer`: Schema, migrations, **transaction contexts**
  - `api-reviewer`: RESTful design, status codes, versioning
  - `architecture-reviewer`: Separation of concerns, coupling
- Severity classification: 🚨 Critical (P0), ⚠️ Important (P1), 💡 Optional (P2)

### Layer 4: Post-Review (5-10 min)
**Purpose**: Share findings and track improvements
- GitHub PR comment posting (inline + summary)
- Review history tracking
- Metrics collection (time saved, issues found)
- Knowledge capture for future reviews

---

## Review Output Guardrails

**CRITICAL**: All reviews MUST follow these constraints for professionalism and readability.

### Length Constraints
- **Target**: <350 lines / 1,750 words for typical PR (<200 LOC)
- **Large PR (>500 LOC)**: <450 lines / 2,250 words maximum
- **Principle**: Concise > Comprehensive when explaining the same thing
- **Quality over quantity**: Focus on proven issues, not speculative concerns
- **Compression tips**:
  - Consolidate related issues into single items
  - Use "Nit:" prefix for P2 items instead of full explanations
  - Reference line numbers instead of showing code blocks when obvious
  - Avoid explaining every edge case - focus on actionable feedback

### Emoji Usage - STRICT LIMITS
**Allowed (Max 10-15 total)**: Use ONLY for status indicators
- ✅ PASS / APPROVE / Success
- ❌ FAIL / BLOCKING / Error
- ⚠️ WARNING / Important issue
- 💡 SUGGESTION / Optional improvement
- 🚨 CRITICAL (for transaction context bugs only)

**PROHIBITED** (Decorative emojis):
- 🎯 🎉 🚀 🔍 📋 💬 🔥 👍 ❤️ 🏆 💪
- No emojis in headings, bullet points, or body text (status only)

### Code Snippets
- **Show full code** only for complex logic (>5 lines) or before/after comparisons
- **Use references** for simple suggestions: "Add nil check at line 42" (not full code block)
- Avoid showing current code + suggested code when brief description suffices

### Redundancy Elimination
- **Explain once** in the appropriate layer
- Brief references elsewhere, no repetition
- Example: Explain infrastructure issues in Layer 1, just reference in summary

### Table Limits
- **Maximum 3 tables** per review
- Use tables for structured data (metrics, test results, file changes)
- Don't use tables for lists that bullets would serve

### Professional Tone
- **Max 5 exclamation marks** in entire review
- No casual language ("awesome!", "super cool", "nailed it")
- No over-praise - be specific and factual
- Balance criticism with recognition of good patterns

### Inline Comments - MANDATORY
**MUST create 3-5 inline comments** for every review covering:
- 1-2 critical/important issues (with line-specific feedback)
- 1-2 good patterns to recognize (balanced feedback)
- 1 suggestion for improvement

**Format**: JSON file with path, line, body
```json
[
  {
    "path": "file.go",
    "line": 42,
    "body": "**⚠️ P1: Issue description**\n\nExplanation and recommendation"
  }
]
```

**Posting**: Use `post_review.sh <pr> <review-file> <action> <inline-comments-file>`

---

## Review Output Template (v4.2)

**MANDATORY**: All reviews MUST follow this structure for consistency and clarity.

```markdown
# Code Review: PR #XXX

## Executive Summary

**Overall Quality**: Excellent | Good | Satisfactory | Needs Improvement | Poor
**Production Readiness**: Ready | Conditional | Not Ready
**Security Risk Level**: None | Low | Medium | High | Critical

**Issue Summary**:
- 🚨 Critical (P0): X issues - Must fix before merge
- ⚠️ Important (P1): X issues - Should fix (may not block merge if core is sound)
- 💡 Optional (P2): X issues - Consider fixing

**Recommendation**: APPROVE | APPROVE WITH CONDITIONS | REQUEST CHANGES
- **If APPROVE WITH CONDITIONS**: Clearly state what conditions must be met (e.g., "Approve core changes if unrelated files split to separate PR")

---

## Layer 0: Intent Alignment

[Validate PR description, confirm intent matches implementation, check for missing context]

## Layer 1: Correctness Gates

[Build validation, test execution, transaction context check, lint results]

## Layer 2: Scope Analysis

### Automation Verification ✅ REQUIRED

**GitHub API (Source of Truth)**:
- Files changed: [X from gh pr view]
- Lines changed: [Y from gh pr view]
- Verification status: ✅ Match / ⚠️ Mismatch detected

**If Mismatch Detected**:
- Local git diff showed: [X files, Y lines]
- Discrepancy: [Explain difference]
- Using GitHub API values (verified source of truth)

### PR Size Metrics

[Analysis using verified metrics from GitHub API]

## Layer 3: Quality Assessment

### Code Quality Patterns
[Analysis of code patterns, design decisions, implementation approach]

### Issues Found

#### 🚨 Critical (P0) - Must Fix
[Critical issues with security, data integrity, or functionality impact]

#### ⚠️ Important (P1) - Should Fix
[Important issues impacting maintainability or performance]

#### 💡 Optional (P2) - Consider Fixing
[Nice-to-have improvements for code quality]

### SOLID Principles Assessment

**Single Responsibility**:
- ✅ Functions have focused purposes
- ⚠️ [Any violations with line references]

**Open/Closed**:
- ✅ Interface-based design allows extension
- ⚠️ [Any violations]

**Liskov Substitution**:
- ✅ All implementations maintain contracts
- ⚠️ [Any violations]

**Interface Segregation**:
- ✅ Interfaces are minimal
- ⚠️ [Any violations]

**Dependency Inversion**:
- ✅ Depends on interfaces, not concrete implementations
- ⚠️ [Any violations]

### Security Assessment

**Risk Level**: None | Low | Medium | High | Critical

**Analysis**:
- ✅ Authentication/authorization checks
- ✅ Input validation
- ✅ Injection prevention
- ⚠️ [Any vulnerabilities or concerns]

**Vulnerabilities**: [List any identified]
**Recommendations**: [Security improvements]

### Positive Observations

✅ **[Pattern/Feature]** (line XX): [Why it's excellent]
✅ **[Pattern/Feature]** (line XX): [Why it's excellent]
✅ **[Pattern/Feature]** (line XX): [Why it's excellent]

---

## Acceptance Criteria for Merge

**Must Have** (Blocking):
- ✅/❌ All tests pass
- ✅/❌ No critical security vulnerabilities
- ✅/❌ [Other blocking items]

**Should Have** (Recommended):
- ✅/⚠️ [Recommended items]

**Nice to Have** (Optional):
- 💡 [Optional improvements]

**Current Status**: X blocking items remaining

---

## Recommendations for Next Steps

### Immediate Actions (Before Merge)
1. [Action with line reference]
2. [Action with line reference]

### Short-term Improvements (Next Sprint)
1. [Improvement suggestion]
2. [Improvement suggestion]

### Long-term Considerations (Future)
1. [Architectural improvement]
2. [Scalability consideration]

### Tooling Integration
- [Linter/tool suggestions]
- [CI/CD recommendations]

### Documentation
- [Doc updates needed]
- [API documentation]

---

## Final Verdict

[Clear statement of approval/rejection with rationale]
```

**Template Usage Notes**:
- **CRITICAL**: Read `references/review-quality-standards.md` for Google's approval decision logic
- Read `references/solid-principles-go.md` for SOLID assessment guidance
- Read `references/security-assessment-checklist.md` for security risk rating
- **Remember**: Favor approval when change improves code health (see Approval Decision Logic section)
- Executive Summary provides instant clarity for stakeholders
- Positive Observations section balances criticism with recognition
- Acceptance Criteria: Separate "Must Have" (P0 blocking) from "Should Have" (P1 non-blocking)
- Structured Next Steps enables prioritization

---

## Automation Scripts (9 Total)

### Core Review Scripts
1. **`analyze_pr_patterns.sh`** - Pattern detection and agent recommendation
   - Detects auth, DB, API, concurrency, **transaction** patterns
   - Suggests which specialized agents to spawn
   - **NEW**: Transaction context detection and warnings

2. **`run_layer1_gates.sh`** - Fail-fast correctness checks
   - Executes build, test, lint in sequence
   - Stops on first failure
   - Provides clear fix guidance

3. **`check_pr_size.sh`** - Scope analysis with GitHub API verification
   - Uses GitHub API as source of truth
   - Verifies against local git diff and warns on mismatches
   - Calculates LOC and file count
   - Recommends split if needed
   - Classifies PR size (small/medium/large)

### GitHub Integration Scripts
4. **`fetch_pr.sh`** - PR retrieval
   - Fetches PR by number or URL
   - Creates review branch
   - Stores PR metadata (number, base branch, URL)

5. **`post_review.sh`** - Review posting with inline comments
   - Posts approval/request-changes/comment reviews
   - Supports inline code comments with line-specific feedback
   - Improved formatting with severity indicators
   - Markdown with collapsible sections

### Helper Scripts
6. **`generate_inline_comments_template.sh`** - Template generator
   - Creates skeleton for inline comments
   - File:line format
   - Ready for agent population

7. **`suggest_pr_split.sh`** - Smart PR splitting
   - Analyzes commit history
   - Suggests logical split points
   - Generates split instructions

8. **`track_review_history.sh`** - Metrics tracking
   - Records review completion
   - Time spent per layer
   - Issues found by severity

9. **`detect_generated_code_issues.sh`** - Generated code detection
   - Identifies auto-generated files
   - Flags issues in generated code
   - Suggests moving to custom code

---

## Reference Documentation (13 Guides)

### Original References (3)
- `google-go-patterns.md` - Google's Go style guide patterns
- `uber-go-patterns.md` - Uber's Go style guide
- `layer3-checklist.md` - Quality review checklist

### Razorpay References (7)
- 🚨 `razorpay-transaction-context.md` - **CRITICAL**: DB transaction patterns (386 lines)
- `razorpay-concurrency-patterns.md` - Goroutines, channels, sync primitives (825 lines)
- `razorpay-error-handling.md` - Error wrapping, sentinel errors (654 lines)
- `razorpay-performance-optimization.md` - Memory, I/O, benchmarking (824 lines)
- `razorpay-testing-best-practices.md` - Table-driven tests, mocking (721 lines)
- `razorpay-idiomatic-go.md` - Packages, structs, interfaces (860 lines)
- `QUICK_REFERENCE.md` - 60-second review checklist (307 lines)

### v4.2 References (2)
- ⭐ `solid-principles-go.md` - SOLID principles adapted for Go with examples and review template (324 lines)
- ⭐ `security-assessment-checklist.md` - Security vulnerability assessment framework with risk ratings (323 lines)

### NEW v4.3 References (1)
- 🎯 `review-quality-standards.md` - Google Engineering Practices code review standards, approval decision logic, and quality metrics (350+ lines)

## Critical New Feature: Transaction Context Detection

### The Problem
Using wrong context inside database transactions is a **CRITICAL** bug that causes:
- Transaction isolation violations
- Incorrect timeout handling
- Context cancellation not propagated
- Difficult-to-debug race conditions
- **Potential data corruption**

### Pattern Detection
```go
// ❌ CRITICAL BUG
func (r *Repo) Update(ctx context.Context) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        payment, err := r.GetPayment(ctx, id)  // Using ctx instead of tctx!
        return r.UpdateStatus(ctx, id, "done")  // WRONG!
    })
}

// ✅ CORRECT
func (r *Repo) Update(ctx context.Context) error {
    return r.db.Transaction(ctx, func(tctx context.Context) error {
        payment, err := r.GetPayment(tctx, id)  // Using tctx!
        return r.UpdateStatus(tctx, id, "done")  // CORRECT!
    })
}
```

### Automatic Detection
When transaction code is detected, `analyze_pr_patterns.sh` now:
1. Flags transaction patterns in diff
2. Spawns `database-reviewer` with CRITICAL warnings
3. Provides specific patterns to check
4. References `razorpay-transaction-context.md` for details

## Usage Examples

### Example 1: Full PR Review
```bash
# Fetch PR
./scripts/fetch_pr.sh 123

# Check size and recommend split if needed
./scripts/check_pr_size.sh main

# Run Layer 1 gates (fail fast)
./scripts/run_layer1_gates.sh

# Analyze patterns and get agent recommendations
./scripts/analyze_pr_patterns.sh main

# Post review to GitHub
./scripts/post_review.sh 123 review-results.md approve
```

### Example 2: Pattern-Based Agent Spawning
```bash
# Get pattern analysis
./scripts/analyze_pr_patterns.sh main

# Output shows detected patterns:
# 🔐 Authentication/Security changes
# 🗄️  Database/Repository changes
# 🚨 Transaction code - CRITICAL: Check transaction context usage!
# 🌐 API/Handler changes

# Recommendation: Spawn 4 specialized agents in parallel
# - backend-engineer:review:security-reviewer
# - backend-engineer:review:database-reviewer (with transaction checks)
# - backend-engineer:review:api-reviewer
# - backend-engineer:review:go-reviewer
```

### Example 3: Transaction Context Review
When transactions are detected, database-reviewer receives:
```
🚨 CRITICAL: Transaction code detected!

MUST check transaction context usage - see references/transaction-context.md

Verify ALL database operations inside Transaction() callbacks use 'tctx' not 'ctx':
  ❌ WRONG: r.GetPayment(ctx, ...) inside Transaction() - uses outer context
  ✅ CORRECT: r.GetPayment(tctx, ...) - uses transaction context

Why this matters:
  • Transaction isolation - operations using 'ctx' execute OUTSIDE the transaction
  • Connection pooling - different contexts may use different DB connections
  • Cancellation propagation - transaction rollback won't cancel 'ctx' operations

Common patterns to flag:
  • Direct calls: r.repo.GetPayment(ctx, id) inside Transaction()
  • Helper calls: r.validateAndSave(ctx, ...) inside Transaction()
  • Nested operations: r.updateStatus(ctx, ...) inside Transaction()
```

## Time Savings

Based on real-world validation:
- **Manual review**: ~33 minutes per PR
- **With v4.0 automation**: ~5 minutes per PR
- **Time saved**: 85% reduction (28 minutes saved)
- **Focus shift**: From finding issues → validating findings

## Integration with Agent Teams

Works seamlessly with specialized review agents:
```
Task: backend-engineer:review:database-reviewer
Prompt: "Review database changes with CRITICAL focus on transaction context..."

Task: backend-engineer:review:security-reviewer
Prompt: "Review authentication and authorization..."

Task: backend-engineer:review:performance-reviewer
Prompt: "Review query performance and indexing..."
```

Run agents **in parallel** by including all Task tool calls in a single message.

## Quality Standards

Based on research from:
- ✅ Google Go Style Guide
- ✅ Uber Go Style Guide
- ✅ Razorpay Best Practices
- ✅ ISO/IEC 25010 Software Quality Model
- ✅ Google Engineering Practices (Code Review Developer Guide)

## Severity Classification

- 🚨 **Critical (P0)**: Must fix - breaks functionality, security, or data integrity
  - Transaction context bugs
  - Security vulnerabilities
  - Data corruption risks
  - Failing tests or build
  - **Note**: Scope issues (PR too large, multiple concerns) are NOT P0 - they're architectural feedback

- ⚠️ **Important (P1)**: Should fix - impacts maintainability, performance, or architecture
  - Missing error wrapping
  - N+1 query problems
  - Race conditions
  - **Scope/architectural concerns**: PR too large (>500 LOC), multiple concerns bundled, unexplained cross-domain changes
  - **Process issues**: Missing documentation, unclear intent
  - **Note**: P1 issues suggest improvements but don't necessarily block well-implemented code

- 💡 **Optional (P2)**: Nice to have - improves code quality
  - Variable naming improvements
  - Missing comments
  - Code organization
  - Speculative improvements (e.g., "consider adding rate limiting" when not proven needed)

---

## Approval Decision Logic (Google Standards)

**The Golden Rule** (from Google Engineering Practices):
> "Reviewers should favor approving a CL once it is in a state where it definitely improves the overall code health of the system being worked on, even if the CL isn't perfect."

### When to APPROVE

✅ **Approve** if the change improves code health, even with minor imperfections:
- Core functionality is correct and well-tested
- No P0 (Critical) issues
- P1/P2 issues exist but don't block the core improvement
- Code is better than what it replaces

✅ **Approve with Conditions** if core is sound but needs non-blocking fixes:
- Core feature is production-ready
- Scope/architectural concerns exist (too large, multiple concerns)
- Suggest: "Approve core X-file change. Consider splitting Y unrelated files to separate PR."
- P1 issues can be addressed in follow-up or current PR

### When to REQUEST CHANGES

❌ **Request Changes** only if:
- P0 (Critical) issues exist: security vulnerabilities, data corruption, transaction bugs, failing tests
- Core functionality is incorrect or incomplete
- Change makes code health worse, not better

### Decision Tree

```
Does this change improve code health?
├─ NO (makes it worse) → REQUEST CHANGES
└─ YES (improves it)
   ├─ Has P0 issues?
   │  ├─ YES → REQUEST CHANGES
   │  └─ NO → Continue
   └─ Has only P1/P2 issues?
      ├─ Core feature is sound → APPROVE WITH CONDITIONS
      │  Example: "Core logic excellent. P1: Split unrelated files."
      └─ Core feature has issues → APPROVE WITH CONDITIONS
         Example: "Good approach. P1: Add error handling before merge."
```

### Handling Scope Issues (CRITICAL)

**Scope issues are NOT merge blockers** - they're architectural feedback:

❌ **Wrong**: "REQUEST CHANGES - PR has 55 files, should be 5"
✅ **Right**: "APPROVE WITH CONDITIONS - Core 5-file change is excellent. Recommend splitting 50 unrelated files to separate PR for clearer review and faster merge."

**Why**: Well-implemented code shouldn't be blocked by bundling decisions. The author can extract core changes or split the PR.

### Balancing Velocity with Quality

**Prioritize**:
1. **Enable progress**: Approve improvements quickly
2. **Maintain quality**: Flag real issues with clear severity
3. **Avoid perfectionism**: Don't block on subjective preferences or speculative issues
4. **Be constructive**: Provide actionable feedback with specific suggestions

**Remember**: Code review is about **continuous improvement**, not **perfection**. Every approved PR should make the codebase better, but not every PR needs to be flawless.

---

## Limitations

- Requires Go codebase
- Best with GitHub integration (scripts use `gh` CLI)
- Pattern detection works on git diffs
- Some checks require buildable code

## Version History

### v4.3 (2026-02-05) - Quality Fixes & Simplicity - CURRENT
- **Google Approval Decision Logic**: "Favor approving a CL once it improves code health"
  - Pragmatic vs perfectionist approach to reviews
  - P2 observations are truly non-actionable (not "should fix")
  - Scope issues treated as feedback, not blockers
  - Context recognition valued (Layer 0 acknowledgment)
- **Conciseness Improvements**: Target 340 lines (was 450+)
  - Avoid over-explaining standard patterns
  - Focus on proven issues, not speculative concerns
  - Eliminate bikeshedding from P1 severity
- **Simplicity Cleanup**: 39% reduction in skill complexity (7,743 → ~4,700 lines)
  - Consolidated duplicate scripts: 11 → 9 scripts
  - Removed old versions: check_pr_size.sh, verify_pr_metrics.sh, post_review_with_comments.sh
  - Simplified pre-review checklist (automatic verification in Layer 2)
- **New Reference**: `review-quality-standards.md` (350+ lines) - Google Engineering Practices alignment
- Total reference documentation: 5,574+ lines across 13 guides

### v4.2 (2026-02-04) - Enterprise-Grade Review Structure
- **Executive Summary**: Quality ratings (Excellent → Poor), production readiness, security risk level, issue counts
- **SOLID Principles Assessment**: Systematic architectural evaluation with Go-adapted checklist
- **Security Risk Rating**: Explicit security assessment (None/Low/Medium/High/Critical) with vulnerability analysis
- **Positive Observations**: Dedicated section for recognizing excellent patterns (balanced feedback)
- **Acceptance Criteria**: Clear merge requirements (Must/Should/Nice to Have)
- **Structured Next Steps**: Prioritized recommendations (Immediate/Short-term/Long-term/Tooling/Documentation)
- **New References**: `solid-principles-go.md` (324 lines), `security-assessment-checklist.md` (323 lines)
- **Inspiration**: Based on comparative analysis with enterprise code review standards
- Total reference documentation: 5,224 lines across 12 guides

### v4.1 (2026-02-04) - Guardrails & Mandatory Inline Comments
- **CRITICAL**: Added back review output guardrails (removed in v4.0 rewrite)
  - Length constraints: <300 lines typical, <400 for large PRs
  - Emoji limits: 10-15 total (status only: ✅ ❌ ⚠️ 💡 🚨)
  - Professional tone requirements
  - Table limits (max 3)
  - Code snippet optimization
- **MANDATORY**: Inline comments now required for every review
  - 3-5 inline comments minimum (2 issues + 1-3 praise/suggestions)
  - Balanced feedback directly on code lines
  - Better UX for PR authors
- Consolidated v3.1 guardrails with v4.0 Razorpay knowledge base

### v4.0 (2026-02-04) - Razorpay Integration
- Added 7 comprehensive reference guides (4,577 lines)
- **CRITICAL**: Transaction context detection and verification
- Enhanced pattern analysis with transaction warnings
- Quick reference guide for 60-second reviews
- Comprehensive coverage of Go best practices

### v3.0 - Phase 2 Automation
- Automated agent spawning based on patterns
- PR splitting assistant
- Inline comments template generator
- Enhanced Makefile parsing
- 85% time reduction validated

### v2.0 - GitHub Integration
- Review history tracking
- Non-blocking execution mode
- Enhanced PR comment posting
- Inline code comments support

### v1.0 - Initial Release
- Five-layer framework
- Basic automation scripts
- Google/Uber reference guides

## Quick Start

1. **Review a PR**:
   ```bash
   ./scripts/fetch_pr.sh <pr-number>
   ./scripts/analyze_pr_patterns.sh main
   ```

2. **Run correctness gates**:
   ```bash
   ./scripts/run_layer1_gates.sh
   ```

3. **Post results**:
   ```bash
   ./scripts/post_review.sh <pr-number> results.md approve
   ```

4. **Check transaction patterns** (if detected):
   - Read `references/razorpay-transaction-context.md`
   - Verify all `Transaction()` callbacks use `tctx` not `ctx`
   - Flag any violations as 🚨 Critical

## Support

- All scripts include `--help` flag
- Reference docs have detailed examples
- Pattern detection provides specific guidance
- Review history tracks metrics over time

---

**v4.3** - Pragmatic code review aligned with Google Engineering Practices - approve improvements, avoid perfectionism.
