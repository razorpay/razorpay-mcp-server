# Performance: Delegated to db-network-optimizer

**This reference file serves as documentation only.**

Performance checks in pre-mortem are handled by delegating to the `db-network-optimizer` skill.

## Architecture

```
pre-mortem (Comprehensive PR Validation)
    ↓
    Step 8: Check Performance
    ↓
    Detects performance-relevant file changes
    ↓
    Invokes: /db-network-optimizer
    ↓
    Receives: Two-section performance report
        1. Duplicate Detection (8 patterns)
        2. Query Optimization (10 checks)
    ↓
    Integrates findings into pre-mortem report by severity
```

## Why Delegation?

✅ **Single Source of Truth**
- All performance check logic lives in db-network-optimizer
- No duplication of pattern definitions
- Easier to maintain and update

✅ **Consistency**
- Same checks whether using pre-mortem or db-network-optimizer directly
- Identical severity levels and fix recommendations
- Users get consistent results

✅ **Reusability**
- db-network-optimizer can be used standalone for deep analysis
- Pre-mortem automatically benefits from improvements to db-network-optimizer
- Both skills stay in sync

## What Pre-Mortem Checks

When PR modifies performance-relevant files:
- Database repositories: `internal/*/repo.go`
- Service clients: `internal/*/service.go`
- API handlers: `internal/*/handler.go`
- Database migrations: `internal/database/migrations/*.sql`
- Background workers: `worker/*`, `internal/job/*`

Pre-mortem invokes:
```bash
/db-network-optimizer "Analyze PR #$PR_NUMBER for performance issues"
```

## Performance Checks Performed

### Duplicate Detection (8 patterns)
1. Duplicate database queries (High)
2. Duplicate service calls (High)
3. N+1 patterns in loops (High)
4. Missing parameter passing (Medium)
5. Missing request cache (Medium)
6. Missing GORM preload (Low)
7. Repeated config fetches (Low)
8. Related entities sequential (Medium)

### Query Optimization (10 checks)
**Level 1 (High Priority):**
1. Missing index on WHERE clause
2. Missing index on JOIN condition
3. Missing composite index

**Level 2 (Medium Priority):**
4. Suboptimal composite index order
5. Missing covering index
6. Index selectivity too low

**Level 3 (Low Priority):**
7. Redundant index (prefix overlap)
8. Unused index (zero usage)
9. Duplicate index (same columns)
10. Query rewrite opportunities

## Integration into Pre-Mortem Report

Performance findings are integrated by severity:

```
🚨 CRITICAL (Must Fix):

1. Missing Index on 10M Row Table
   Location: internal/payments/repo.go:42
   Issue: No index on (merchant_id, status) → Full table scan
   Source: db-network-optimizer

   Fix: CREATE INDEX idx_payments_merchant_status
        ON payments(merchant_id, status);

   Performance: 500ms → 5ms (100× faster)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ HIGH (Should Fix):

2. Duplicate Service Calls (6×)
   Location: internal/payer/core.go:55, internal/forex/core.go:69, +4
   Issue: FetchOrderDetails() called 6 times with same orderId
   Source: db-network-optimizer
   Impact: 120-300ms wasted per request

   Fix: Pass order object through call chain
   [Code example from db-network-optimizer]

   Improvement: 85-95% reduction in service calls
```

## For Deep Performance Analysis

Users can run db-network-optimizer directly for:
- More detailed two-section report
- Comprehensive pattern catalog
- Multiple fix strategies with code examples
- Full repository analysis (--full flag)

```bash
# Deep dive on same PR
/db-network-optimizer "Review PR #456"

# Full repository audit
/db-network-optimizer --full
```

## Reference

See `db-network-optimizer` skill for:
- Complete pattern catalog
- Detection methods (AST patterns, flow analysis)
- Fix strategies with before/after examples
- Validation results
- Technical implementation details
