---
name: project-planner
description: Creates detailed, realistic project plans from design documents or specifications. Extracts tasks, estimates effort with AI-assisted coding factored in, maps dependencies, identifies parallelization, and performs critical path analysis. Use when the user asks to create a project plan, task breakdown, sprint plan, or effort estimate from a spec, RFC, design doc, or plan document.
---

# Project Planner

Creates realistic, AI-adjusted project plans from design documents. Follows a strict progressive process: gather context, ask clarifying questions, draft, then iteratively refine with user feedback.

**Core philosophy:** Never assume. Always ask. Humans shape the plan -- the skill provides structure and rigor.

## Usage

Invoke when a user provides a design document, RFC, technical spec, or plan document and asks for:
- A project plan or task breakdown
- Effort estimates (with or without AI-adjustment)
- Sprint planning or milestone decomposition
- Critical path analysis or parallelization strategy
- Timeline projections for different team sizes

The skill follows a strict 4-phase workflow: **Gather Context -> Draft Plan -> Present & Iterate -> Timeline Analysis**. It will ask clarifying questions before producing any output.

## Examples

**Input:** "Here's our design doc for the new payment gateway. Create a project plan."

**Output:** A structured task list with columns for Task ID, Title, Details, Milestone, Dependencies, Effort (Days), Priority, Parallel Group, Delegatable, and Notes -- plus critical path analysis and timeline projections for 1-N developers.

**Input:** "Break down this RFC into sprint tasks. We use Cursor and deploy on merge to devstack."

**Output:** AI-adjusted effort estimates grouped by milestone, with deployment tasks absorbed into "Review, Merge & Deploy" entries (no separate deploy tasks), and a parallelization sweet-spot analysis.

## Phase 1: Gather Context

### Step 1: Read the Source Document

Read the full design document the user provides. Do not skim. Extract:
- Functional requirements and their implementation details
- Non-functional requirements (performance, security, reliability)
- Architecture and component boundaries
- Infrastructure needs (queues, databases, deployments, IAM)
- Integration points between components
- Testing requirements (unit, integration, E2E, performance)
- Rollout/migration strategy
- Monitoring and observability needs
- Documentation and handoff requirements

### Step 2: Ask Mandatory Clarifying Questions

**Do NOT start planning until these are answered.** Use the AskQuestion tool if available, otherwise ask conversationally. Ask in batches, not all at once.

**Batch 1 -- Effort Model:**
- What is the effort unit? (e.g., "1 day = 3 hours of coding", "1 day = full workday")
- Will AI-assisted coding tools (Cursor, Claude Code, Copilot) be used?
- Should estimates reflect AI-adjusted effort or traditional effort?

**Batch 2 -- Team & Process:**
- Single developer or multi-dev? If multi, how many?
- What is the code review process? (async PR review, pair programming, sync review)
- Are there external team dependencies? (e.g., DevOps, platform team, security review)

**Batch 3 -- Environments & Deployment:**
- What are the environment names and promotion path? (e.g., devstack -> staging -> production, or devstack -> production)
- Does each feature deploy to lower env on merge, or is there a separate deployment step?
- Is there a feature flag / rollout system?

**Batch 4 -- Output Preferences:**
- Preferred output format? (CSV, markdown table, spreadsheet)
- What columns are needed? Suggest defaults: Task ID, Title, Details, Milestone, Dependencies, Effort (Days), Priority, Parallel Group, Delegatable, Notes
- How many months does the user expect to calculate for? (used to convert days to months, default: 22 working days/month)

**Batch 5 -- Scope (ask if unclear from the document):**
- Which phases from the document should be included? (all, or specific milestones)
- Should operational tasks (monitoring, docs, KT) be included?
- Where should the output file be saved?

### Step 3: Confirm Understanding

Before drafting, summarize back to the user:
- Total scope (what milestones are included)
- Effort unit and AI-adjustment approach
- Team model (single dev, with parallelization shown)
- Environment names and deployment flow
- Output format and location

Get explicit confirmation before proceeding.

## Phase 2: Draft the Plan

### Task Extraction Rules

1. **Read the spec exhaustively.** Every requirement, every component, every infrastructure need becomes one or more tasks.
2. **Group by milestone.** Each logical deliverable (package, module, integration, rollout phase) is a milestone.
3. **Number hierarchically.** Use M.T format (e.g., 1.1, 1.2, 2.1). No milestone header rows in the output -- the Milestone column provides that context.

### Task Granularity Rules (Critical)

Apply these checks to every task. If a task violates a rule, merge it with its natural neighbor.

| Rule | Threshold | Action |
|------|-----------|--------|
| **Minimum viable task** | < 2 hours with AI | Merge into the task it naturally precedes or follows |
| **Config/boilerplate tasks** | Struct definitions, TOML entries, constants | Absorb into the implementation task that uses them |
| **Registration/wiring tasks** | "Register handler", "Wire into boot", "Add to startup" | Absorb into the implementation task |
| **Pure interface/design tasks** | When spec already has the design | Reduce effort or merge into implementation |
| **Separate unit test tasks** | When AI generates tests in minutes | Absorb into the implementation task ("+Tests" in title) |
| **Separate "Merge PR" tasks** | Merge is part of review | Combine into "Review, Merge & Deploy" (1 task, max 1 day) |

### Effort Estimation Rules

1. **AI-adjusted by default.** If user confirms AI tools, reduce estimates for:
   - Boilerplate/scaffolding: 70-80% reduction
   - Test generation: 70-80% reduction
   - Code comprehension/ramp-up: 50% reduction
   - Design (when spec is detailed): 50% reduction
   - Complex algorithmic logic: 10-20% reduction only
   - Integration/debugging: minimal reduction (still needs human judgment)
   - Production rollout/monitoring: no reduction (wall-clock time)

2. **Round to whole numbers.** Minimum effort = 1 day. No 0.5 day tasks.
   - 0.1-0.6 days -> 1 day (round up to minimum)
   - 0.7-0.9 days -> 1 day
   - 1.1-1.4 days -> 1 day
   - 1.5-1.9 days -> 2 days
   - General: round to nearest whole number, minimum 1.

3. **Review/Merge/Deploy = 1 day max.** Code review is async -- the developer moves to the next task while review happens. Never separate "merge" or "deploy" as standalone tasks.

4. **Bug fix buffers are real.** After testing phases, include 1-2 day buffers. These are not padding -- they're realistic allowance for integration issues.

5. **Production rollout = wall-clock time.** Ramp-up phases (5% -> 25% -> 50% -> 100%) reflect monitoring periods, not coding effort. Do not AI-adjust these.

### Dependency Rules

1. **Relax dependencies aggressively.** If Task B only needs Task A's *interface* (not full implementation), depend on the interface/design task, not the merge task.
2. **Identify fork points.** After a task completes, note which downstream tasks can run in parallel. Add a "FORK POINT" note.
3. **Mark convergence points.** Where multiple streams must merge (e.g., integration testing needs all features), note the dependencies explicitly.

### Deployment Flow Rules

1. **No separate "Deploy to [env]" tasks** if each feature deploys on merge. The "Review, Merge & Deploy" task handles this.
2. **No "Deploy Full Stack" task** if individual feature deployments already achieve this. By the time all features are merged, the full stack is deployed.
3. **Infrastructure provisioning** (Terraform, K8s, IAM) must happen before the code that uses it can be deployed. Place infrastructure tasks accordingly.
4. **Production infrastructure** provisioning happens after lower-env testing passes, before production rollout.

### Monitoring & Operational Tasks

1. **Metrics instrumentation** (Prometheus counters, histograms) is code -- include with implementation tasks.
2. **Dashboards and alerts** are post-production -- schedule after the system is live and producing real data.
3. **Runbooks, documentation, KT** are post-rampup operational tasks.

### Structural Columns

Generate these columns (adjust based on user preference):

| Column | Purpose |
|--------|---------|
| Task ID | Hierarchical: M.T (e.g., 1.1, 2.3) |
| Task Title | Concise, action-oriented |
| Task Details | What specifically to implement (file names, functions, patterns) |
| Milestone | Logical deliverable this task contributes to |
| Dependencies | Comma-separated Task IDs, or "None" |
| Effort (Days) | AI-adjusted, whole numbers, minimum 1 |
| Priority | P0 (critical path), P1 (important), P2 (can defer) |
| Parallel Group | Tag for tasks that can run simultaneously (e.g., PG-1.2) |
| Delegatable | Yes/No -- can a second dev pick this up independently? |
| Notes | Fork points, dependency relaxation notes, AI-assist notes |

## Phase 3: Present and Iterate

### First Presentation

After generating the draft, always present ALL of the following:

1. **Plan summary:** total task count and total effort (days and months)
2. **Per-milestone effort breakdown** as a table
3. **Critical path:** list the longest dependency chain with its total duration
4. **Timeline by dev count:** how long the project takes with 1, 2, 3, and N devs (see Phase 4)
5. **Parallelization sweet spot:** the dev count beyond which adding devs provides no further compression
6. **Parallel tracks:** which milestone streams can run concurrently and what each dev would work on

Then ask: "Does this structure match your expectations? Any tasks that seem inflated, too granular, or missing?"

### Iteration Protocol

For each round of feedback:
1. Understand the specific concern (over-estimation? wrong grouping? wrong sequencing?)
2. Make the change
3. Show what changed and the new totals
4. Ask if there's more to adjust

**Never push back on the user's domain knowledge.** If they say "this task takes 30 minutes with AI," believe them and merge it.

### Final Delivery

After the user approves the plan:
1. Write the output file to the agreed location
2. Provide the **full summary** (same as First Presentation: totals, milestones, critical path, timeline by dev count, sweet spot, parallel tracks)
3. If the plan changed since first presentation, recompute all analysis

## Phase 4: Timeline & Parallelization Analysis (Mandatory)

This analysis is a **mandatory part of every plan delivery** -- both in the first presentation and the final delivery. Never skip it.

### Step 1: Compute Earliest Completion (EC)

For every task, calculate: `EC = max(EC of all dependencies) + task effort`.
Tasks with no dependencies: `EC = effort`.

### Step 2: Identify the Critical Path

Trace backward from the final task: at each step, follow the dependency with the highest EC. This chain is the critical path. Present it as:

```
Critical Path (X days):
Task A (Xd) -> Task B (Xd) -> ... -> Final Task (Xd)
```

### Step 3: Calculate Slack

For non-critical-path tasks: `Slack = EC(next critical task) - EC(this task)`.
Call out tracks with significant slack -- these are where parallel devs add value.

### Step 4: Schedule for N Devs

For each dev count (1, 2, 3, and the sweet spot N):
1. Assign tasks to devs greedily, prioritizing critical path tasks
2. Respect all dependency constraints
3. Track calendar day per dev
4. Note what each dev works on at a high level

### Step 5: Present the Timeline Table

Always present this table:

```
| Devs | Calendar Days | Months (~22d/mo) | Notes |
|------|--------------|------------------|-------|
| 1    | X            | Y                | Everything sequential |
| 2    | X            | Y                | [which tracks run parallel] |
| 3    | X            | Y                | [sweet spot or additional parallel tracks] |
| N    | X            | Y                | Theoretical minimum (critical path) |
```

Include:
- **Sweet spot callout:** "N devs is the sweet spot -- adding more provides no further compression because the bottleneck is [describe the irreducible sequential chain]."
- **Per-dev schedule overview:** For the recommended dev count, show what each dev works on in which time window (e.g., "Dev 1: M1 (days 0-7) -> M3 (days 7-12) -> ...")
- **Where the time goes:** Break down the critical path into logical phases (build X days, test Y days, rollout Z days)

## Anti-Patterns Checklist

Before finalizing, verify the plan does NOT contain:

- [ ] Tasks under 2 hours of AI-assisted work as standalone entries
- [ ] Separate "unit test" tasks (should be absorbed into implementation)
- [ ] Separate "merge PR" tasks (should be part of "Review, Merge & Deploy")
- [ ] Separate "deploy to [env]" tasks when deploys happen on merge
- [ ] "Full stack deployment" task when individual merges already achieve this
- [ ] Config struct / boilerplate as standalone tasks
- [ ] "Register handler" / "wire into boot" as standalone tasks
- [ ] Monitoring dashboards scheduled before production rollout
- [ ] 0.5 day or fractional effort values
- [ ] Assumptions about deployment model without asking the user
- [ ] Assumptions about environment names without asking the user
- [ ] Assumptions about team size without asking the user

For detailed anti-patterns and examples, see [references/reference.md](references/reference.md).
