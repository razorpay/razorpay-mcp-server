# Project Planner Reference

Detailed anti-patterns, examples, and lessons learned from real-world project planning sessions. Read this when you need to resolve ambiguity or justify a decision.

## Anti-Pattern Deep Dives

### 1. The "Boot Init" Trap

**Problem:** Creating a standalone task for "Add X to application startup" or "Wire X into boot file."

**Why it's wrong:** Adding a function call to `main.go` or a boot sequence is 5-15 minutes of work. It's not a plannable unit of work.

**Fix:** Absorb into the implementation task that creates the thing being wired up. Example:
- Bad: "Implement SWR Package" (2d) + "Wire SWR into Boot" (1d) = 3d, 2 tasks
- Good: "Implement SWR Package + Boot Integration" (2d) = 2d, 1 task

### 2. The "Config Struct" Trap

**Problem:** Separate tasks for "Create config structs" or "Add TOML entries."

**Why it's wrong:** Config structs are written as part of the implementation. You don't write a config struct without the code that uses it.

**Fix:** Absorb into the implementation task. Mention config in the task details.

### 3. The "Context Propagation" Trap

**Problem:** Standalone task for `SetXInContext()` / `GetXFromContext()` -- literally 2 functions with a typed context key.

**Why it's wrong:** 15 minutes of work. Not a plannable unit.

**Fix:** Merge with the feature that uses the context (e.g., attribute fetching + context propagation = one task).

### 4. The "Separate Deploy" Trap

**Problem:** Tasks like "Deploy to DevStack" or "Deploy Full Stack to Staging" as standalone entries.

**Why it's wrong:** If each feature deploys to the lower environment when its PR is merged, there's no separate deploy step. By the time all features are merged individually, the full stack is already deployed.

**Fix:**
- Each "Review, Merge & Deploy" task includes deployment to the lower env.
- No standalone "Deploy Full Stack" task -- it's already there from individual deploys.
- Only have a standalone deploy task for environments that require explicit provisioning (e.g., production infrastructure).

### 5. The "Monitoring Before Production" Trap

**Problem:** Scheduling Grafana dashboards, CloudWatch alarms, and alerting rules before the system is in production.

**Why it's wrong:** Dashboards and alerts are configured based on observed production behavior. Creating them before production means guessing at thresholds and missing real patterns.

**Fix:**
- Metrics instrumentation (Prometheus counters/histograms) IS code -- include with implementation tasks.
- Grafana dashboards and alert rules are POST-PRODUCTION tasks -- schedule after full rollout.
- The team observes real production behavior first, then configures dashboards and alerts.

### 6. The "Inflated Ramp-up" Trap

**Problem:** "Codebase Ramp-up & Context Building" estimated at 3-5 days.

**Why it's wrong:** With AI-assisted code comprehension and a detailed design document that already analyzes the codebase, ramp-up is 1 day (one 3-hour session of reading code with AI explaining patterns).

**Fix:** 1 day for ramp-up when: (a) AI tools are available, AND (b) a detailed spec/design doc exists.

### 7. The "Inflated Design" Trap

**Problem:** "Design & Finalize Interfaces" estimated at 3-5 days.

**Why it's wrong:** If the design document already specifies interfaces, function signatures, and data structures in detail, the "design" task is really "transcribe the spec into code."

**Fix:** 1 day when the spec is detailed. 2 days if significant decisions remain.

## Merging Decision Matrix

When deciding whether to merge two tasks:

| Scenario | Merge? | Resulting Effort |
|----------|--------|-----------------|
| Task A (1d) + Task B depends on A (1d), B < 2hr work | Yes | Keep A's effort |
| Task A (2d) + Task B is boilerplate for A (0.5d) | Yes | Keep A's effort |
| Task A (1d) + Task B shares same files as A (1d) | Yes | Sum if > 1d, else 1d |
| Task A (2d) + Task B is complex and independent (2d) | No | Keep separate |
| Two test tasks for the same module | Yes | 1d combined |
| "Code Review" + "Merge PR" + "Deploy" | Always merge | 1d combined |
| Feature flag + config entries for same feature | Yes | 1d combined |
| Interface design + implementation (spec is detailed) | Maybe | Merge if design < 1hr |

## AI-Adjustment Multipliers

These are guidelines, not formulas. Adjust based on spec detail and codebase complexity.

| Task Type | Traditional | AI-Adjusted | Multiplier | Rationale |
|-----------|------------|-------------|------------|-----------|
| Boilerplate/scaffolding | 2d | 1d | 0.5x | AI generates from description |
| Unit test generation | 2d | 1d | 0.5x | AI generates from spec + code |
| Code comprehension | 2d | 1d | 0.5x | AI explains code patterns |
| Design (spec detailed) | 2d | 1d | 0.5x | Spec already has the design |
| Design (spec vague) | 3d | 2d | 0.7x | Human decisions still needed |
| Complex algorithms | 3d | 2-3d | 0.8x | AI helps but human validates |
| Integration/wiring | 2d | 2d | 1.0x | Requires understanding of both sides |
| Debugging/bug fixes | 2d | 2d | 1.0x | Root cause analysis is human work |
| Infra provisioning | 2d | 2d | 1.0x | Terraform apply, verify, troubleshoot |
| Production rollout | 2d | 2d | 1.0x | Monitoring periods are wall-clock |
| External coordination | 1d | 1d | 1.0x | Waiting on other teams |

## Critical Path Analysis Method

### Step 1: Compute Earliest Completion (EC)

For each task, EC = max(EC of all dependencies) + task effort.
Tasks with no dependencies: EC = effort.

### Step 2: Identify Critical Path

Trace backward from the final task: at each step, follow the dependency with the highest EC. This chain is the critical path.

### Step 3: Calculate Slack

For each non-critical task: Slack = EC(first downstream critical task that depends on it) - EC(this task).
Tasks on the critical path have 0 slack.

### Step 4: Schedule for N Devs

Greedy algorithm:
1. Maintain a priority queue of "ready" tasks (all dependencies complete)
2. Assign tasks to free devs, prioritizing critical path tasks
3. When a dev finishes, mark task complete, add newly-ready tasks to queue
4. Track calendar day for each dev

### Step 5: Find the Sweet Spot

Increase N until adding another dev does NOT reduce calendar time. That N is the sweet spot -- the bottleneck is the critical path's irreducible sequential chain.

## Milestone Ordering Guidelines

The typical ordering for a backend feature with infrastructure:

```
1. Foundation packages (generic, reusable)
2. Integration packages (project-specific wiring)
3. Feature testing in lower env (isolated, per-feature)
4. Supporting packages (parallel track: access tracking, queues, etc.)
5. Infrastructure provisioning for lower env
6. Worker/background job infrastructure
7. Webhook/trigger integration
8. Legacy cleanup (code changes)
9. Full-stack testing in lower env (already deployed via individual merges)
10. Production infrastructure provisioning
11. Production rollout (shadow -> percentage ramp -> full)
--- POST RAMPUP ---
12. Monitoring dashboards & alerts (observe real behavior)
13. Legacy cleanup in production
14. Documentation & operational handoff
```

Adapt based on the project. The key principle: **build -> deploy per feature -> test isolated -> full-stack test -> prod provision -> prod rollout -> observe -> operational**.

## Questions to Ask When the User Says "This Seems Inflated"

1. "How detailed is the spec for this task?" (detailed spec = lower estimate)
2. "Will AI be generating the boilerplate?" (yes = lower estimate)
3. "Is this pure code, or does it involve external coordination?" (external = keep estimate)
4. "Can this be absorbed into an adjacent task?" (often yes)
5. "Is this task doing one thing or two?" (if two, maybe keep separate but re-estimate)
