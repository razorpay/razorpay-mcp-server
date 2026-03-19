---
name: tech-spec-reviewer
description: Comprehensive technical specification reviewer that analyzes tech specs for alternative approaches, edge cases, optimizations, testing strategy, observability, NFRs (performance, security, availability), data consistency, and error handling. Use when asked to review tech specs, design docs, architecture proposals, or when user says "review this tech spec" or "check if I'm missing anything in this design".
disable-model-invocation: true
---

# Tech Spec Reviewer

Conduct comprehensive reviews of technical specifications to identify gaps, risks, and improvement opportunities before implementation.

## Review Workflow

When reviewing a tech spec, follow this systematic approach:

### 1. Understand the Spec

First, thoroughly read and understand the technical specification:

- Identify the problem being solved
- Understand the proposed solution and architecture
- Note the scope and constraints
- Identify the domain (backend services, distributed systems, data pipelines, etc.)

### 2. Load Review Checklist

Read the comprehensive review checklist to understand what to look for:

```bash
references/review-checklist.md
```

This checklist covers 10 major categories:
1. Alternative Approaches & Trade-offs
2. Edge Cases & Failure Scenarios
3. Optimizations
4. Testing Strategy
5. Observability & Metrics
6. Non-Functional Requirements (NFRs)
7. Data Consistency & Correctness
8. Error Handling & Resilience
9. Rollout & Rollback Plan
10. Dependencies & Integration

### 3. Identify Domain-Specific Edge Cases

Based on the domain of the system, consult the edge cases catalog:

```bash
references/edge-cases-catalog.md
```

This catalog organizes common edge cases by domain:
- Payment Systems
- Distributed Systems
- Database Systems
- API & HTTP
- Async Processing & Queues
- Data Migration
- Authentication & Authorization
- And more...

Focus on edge cases relevant to the spec's domain.

### 4. Conduct the Review

For each category in the checklist, evaluate the tech spec:

#### For Each Category, Ask:
1. **Is it addressed?** Is the category covered in the spec?
2. **Is it specific?** Are there concrete details, not just "TBD" or "NA"?
3. **Is it complete?** Are important aspects missing?
4. **Is it correct?** Are there flaws in the approach?

#### Assign Severity Levels:
- 🔴 **CRITICAL**: Must fix before implementation - high risk of production issues, data loss, or security breach
- 🟠 **HIGH**: Should fix before implementation - likely to cause issues
- 🟡 **MEDIUM**: Should address during implementation - may cause issues
- 🔵 **LOW**: Consider for improvement - nice to have

#### Flag Red Flags:
- "To be added" / "TBD" for critical sections (NFRs, testing, monitoring, rollback)
- "NA" for important sections without justification (especially security, reliability)
- Only one approach considered without alternatives
- Generic statements without specifics ("high availability" without SLA)
- Missing error handling or failure scenarios
- No performance targets or capacity planning

### 5. Structure the Review Output

Follow the standard review output format defined in:

```bash
references/review-output-format.md
```

The review should include:

1. **Executive Summary** (3-5 bullets)
   - Overall assessment
   - Critical blockers
   - Key recommendations

2. **Detailed Findings by Category**
   - Use consistent severity levels and formatting
   - Be specific and reference section numbers
   - Explain the risk/impact for each issue
   - Provide actionable recommendations

3. **Recommendations Summary**
   - Priority 1: Must fix before implementation
   - Priority 2: Should fix during implementation
   - Priority 3: Consider for future

4. **Questions for Authors**
   - Clarifying questions that need answers

### 6. Review Principles

**Be Constructive and Specific:**
- ❌ "Testing is weak"
- ✅ "No load testing plan defined (Section 10.3). Recommend adding load test with target of 10 TPS to validate performance under expected load."

**Balance Criticism with Recognition:**
- Point out strengths and well-done aspects
- Helps authors understand what's working

**Consider Context:**
- KTLO (Keep The Lights On) products have different standards than strategic products
- Adjust expectations based on scope and constraints
- Acknowledge trade-offs explicitly made

**Focus on Risk:**
- Explain WHY something is important
- What could go wrong if not addressed?
- Quantify impact when possible

**Provide Detailed Examples - CRITICAL:**
- **For Edge Cases**: Provide concrete scenarios that show the failure path
  - Show the sequence of events that leads to the problem
  - Include example data/state that triggers the edge case
  - Explain the user impact or system behavior
- **For Optimizations**: Show before/after scenarios with measurable improvements
  - Include performance numbers where possible
  - Show code snippets or architectural diagrams if helpful
  - Explain the trade-offs of implementing the optimization
- **Always give developers enough detail** to understand the issue and make an informed decision

**Ask Questions:**
- If something is unclear, ask for clarification
- Don't assume intent

## Common Review Patterns

### Payment/Financial Systems

**Extra scrutiny on:**
- Data consistency and money loss prevention
- Idempotency for all transaction operations
- Concurrency control for balance updates
- Audit trail and compliance
- Security (PII, encryption, access control)
- Rollback safety (no money loss during rollback)

### Distributed Systems

**Extra scrutiny on:**
- Network partition handling
- Consistency guarantees (eventual vs strong)
- Distributed transactions and compensating transactions
- Service discovery and failover
- Circuit breakers and timeouts
- Message ordering and duplicate handling

### Data Migration

**Extra scrutiny on:**
- Data validation and reconciliation strategy
- Dual-write consistency
- Rollback with data safety
- Zero-downtime migration approach
- Merchant/user migration phasing
- Backfill verification

### High-Traffic Systems

**Extra scrutiny on:**
- Performance targets (TPS, latency)
- Load testing plan
- Auto-scaling strategy
- Database indexing and query optimization
- Caching strategy
- Rate limiting

## Example Review

For a concrete example of a complete tech spec review following this workflow, see the example in `references/review-output-format.md` which demonstrates:

- How to structure findings by category
- How to assign severity levels
- How to write specific, actionable recommendations
- How to balance criticism with recognition
- How to format the executive summary and recommendations

## Explaining Edge Cases and Optimizations with Examples

**CRITICAL REQUIREMENT**: When identifying edge cases or optimizations, you MUST provide detailed examples that help developers understand and make decisions.

### For Edge Cases - Provide Concrete Scenarios

Always include:
1. **Scenario**: Step-by-step description of the failure path
2. **Example**: Concrete data/state that triggers the issue
3. **Impact**: What breaks - user experience, data corruption, money loss
4. **Recommendation**: How to handle it with code snippets or design changes

**Format:**
```
Edge Case: [Concise name]
Scenario:
  1. [Initial state]
  2. [Action A happens]
  3. [Action B happens]
  4. [Conflict/failure occurs]

Example: [Concrete data showing the problem]

Impact: [What goes wrong - be specific]

Recommendation: [Solution with code/SQL if applicable]
```

### For Optimizations - Provide Before/After Analysis

Always include:
1. **Current Approach**: What the spec proposes
2. **Problem**: Why it's suboptimal (with numbers)
3. **Proposed Approach**: Alternative with explanation
4. **Impact**: Measurable improvements (latency, cost, throughput)
5. **Trade-offs**: Complexity, development time, operational burden
6. **Recommendation**: When to implement (always, only if X TPS, skip if Y)

**Format:**
```
Optimization: [Concise name]
Current Approach: [What spec says]
Problem: [Why suboptimal with metrics - "Query takes 200ms at 100K rows"]
Proposed Approach: [Alternative]
Impact: [Measurable - "Reduce latency from 200ms to 5ms (40x improvement)"]
Trade-offs: [What we sacrifice - "Adds 100MB disk space, slightly slower writes"]
Recommendation: [Clear decision guidance - "MUST implement if TPS > 100, SKIP if KTLO product"]
```

### Why Detailed Examples Matter

- **Developers can assess severity**: "200ms query at 100K rows" vs vague "slow query"
- **Enables informed decisions**: Trade-offs let developers choose based on their context
- **Saves round-trip questions**: Complete examples reduce back-and-forth
- **Educates the team**: Examples teach patterns for future designs

### When to Provide Examples

**Always provide detailed examples for:**
- Race conditions and concurrency issues
- Idempotency and retry logic
- Database query optimization (show query plans, execution times)
- Caching strategies (hit ratios, latency improvements)
- Cost optimizations (show dollar amounts and projections)
- Security vulnerabilities (show attack vectors)
- Data loss scenarios (show the data flow)

**Keep it concise for:**
- Well-known patterns (e.g., "Add circuit breaker" - no need to explain circuit breakers)
- Low-impact suggestions
- Obvious fixes

## Tips for Effective Reviews

1. **Read the entire spec first** before starting detailed review - understand context

2. **Use the checklists systematically** - don't rely on memory alone

3. **Focus on high-impact issues** - not every issue is critical

4. **Be specific with recommendations** - reference sections, provide examples

5. **Consider the audience** - senior engineers need different detail than junior engineers

6. **Validate assumptions** - ask questions when something is unclear

7. **Check for consistency** - do different sections contradict each other?

8. **Think like an attacker** - what could go wrong? What edge cases would break this?

9. **Think about operations** - how will this be debugged, monitored, rolled back?

10. **Consider the timeline** - is "TBD" acceptable given the implementation schedule?

11. **Provide actionable examples** - especially for edge cases and optimizations (see above)

## Common Gaps to Watch For

- **Migration without validation strategy** - how do we know it worked?
- **No rollback plan** - what if we need to revert?
- **Generic NFRs** - "highly available" instead of "99.9% SLA"
- **Missing edge cases** - especially concurrency and failure scenarios
- **No monitoring** - how will we detect issues?
- **Security as afterthought** - marked "NA" or "TBD"
- **Testing plan incomplete** - no load testing, no migration testing
- **Error handling unspecified** - what error does user see?
- **Timeout strategy missing** - every external call needs timeout
- **No capacity planning** - will it handle the load?
