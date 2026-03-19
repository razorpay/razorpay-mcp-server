---
name: semgrep-devrev-remediation
description: |
  Remediates SCA (Software Composition Analysis) vulnerable dependencies end-to-end using DevRev tickets.
  DevRev SCA tickets already contain Semgrep finding data (CVE, safe version, remediation guide),
  so no separate Semgrep MCP server is needed.

  Executes an 8-step workflow: extract remediation from DevRev ticket, identify stable versions,
  gather changelogs, scan repo usage, analyze breaking changes, write security tests, run tests,
  and create a PR or resolution plan.

  Use when:
  - Remediating vulnerable npm dependencies tracked in DevRev SCA tickets
  - Applying npm overrides for transitive dependency vulnerabilities
  - Writing security verification tests for dependency fixes
  - Creating PRs for dependency security patches

  Triggers: "remediate dependencies", "fix vulnerable dependency", "sca remediation",
  "dependency upgrade", "npm override vulnerability", "devrev sca fix",
  "devrev security ticket remediation"

  Required: DevRev MCP server or DevRev PAT token (for API calls)
  Required CLI tools: npm, git, gh (GitHub CLI), curl
user-invocable: true
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, WebFetch
---

# Semgrep DevRev Remediation Skill

Remediates vulnerable npm dependencies tracked via DevRev SCA tickets. DevRev tickets already contain all necessary Semgrep finding data (CVE, minimum safe version, remediation guidance), so this workflow operates with **DevRev as the single source of truth**.

## How It Works

```
DevRev SCA Ticket
  ├── title: package name + vulnerability
  ├── body: CVE details, remediation steps
  └── custom_fields: Semgrep URL, safe version, rule ID
         |
         v
  8-Step Remediation Workflow
         |
         v
  PR (if tests pass) or Resolution Plan (if tests fail)
```

## Prerequisites

- DevRev MCP server configured OR DevRev PAT token for direct API calls (see [references/mcp-config.md](references/mcp-config.md))
- npm project with `package.json` and `package-lock.json`
- GitHub CLI (`gh`) authenticated for PR creation
- Git repository with push access

## The 8-Step Workflow

For each vulnerable dependency (or logical group of related dependencies):

| Step | Action | Input |
|------|--------|-------|
| 1 | Extract remediation from DevRev ticket | Ticket body + custom_fields |
| 2 | Query npm registry for stable target versions | Package name |
| 3 | Gather release notes / changelogs | npm registry + GitHub |
| 4 | Scan repo for direct usage | grep source code |
| 5 | Breaking change analysis | Semver delta + usage map |
| 6 | Write security verification tests | Test templates |
| 7 | Run tests (security + build + existing) | npm/ng commands |
| 8 | Create PR or resolution plan | git + gh CLI |

See [references/remediation-workflow.md](references/remediation-workflow.md) for the complete workflow with commands.

## Quick Start

### 1. Fetch and Group DevRev Tickets

Retrieve all SCA tickets for the target repository from DevRev, then group by vulnerable dependency. See [references/devrev-ticket-fetching.md](references/devrev-ticket-fetching.md).

### 2. Prioritize Dependency Groups

| Priority | Action | SLA |
|----------|--------|-----|
| P0 Critical | Immediate remediation | Same day |
| P1 High | Next sprint | 1 week |
| P2 Medium | Backlog | 2 weeks |
| P3 Low | Track | 1 month |

### 3. Execute 8-Step Workflow Per Group

One PR per dependency group. Each group fixes one root vulnerability across all related tickets.

### 4. Determine Fix Strategy

| Dependency Type | Fix Strategy | Reference |
|----------------|--------------|-----------|
| Transitive dev dep | npm override | [npm-override-patterns.md](references/npm-override-patterns.md) |
| Direct dependency | Version bump in `package.json` | Standard upgrade |
| Nested transitive | Override + verify lockfile | [npm-override-patterns.md](references/npm-override-patterns.md) |

## Dependency Grouping Strategy

Group related dependencies that share a root vulnerability:

```
Example: CVE-2023-45133 (@babel/traverse)
  Group A: All 6 packages whose transitive chain includes @babel/traverse
  Fix: Single override for @babel/traverse -> resolves all 6 tickets
```

## Test Strategy

Security verification tests validate that:
1. The override/upgrade exists in `package.json`
2. The resolved version in `package-lock.json` meets the safe threshold
3. No nested copies of the vulnerable package remain at unsafe versions
4. Existing overrides are preserved (no regressions)
5. The lockfile structure is valid

See [references/security-test-patterns.md](references/security-test-patterns.md) for test templates.

## PR Convention

Branch naming: `fix/semgrep-{dependency-name}`
Commit message format:

```
fix(security): add {package} override to resolve {CVE-ID}

Add npm override for {package} >= {version} to mitigate {CVE-ID}
({severity} - {vulnerability description}).

Root cause: {explanation of transitive dependency chain}
Fix: {description of the override or upgrade}

Resolves {N} DevRev tickets: {ISS-IDs}

Verification:
- Security tests: {pass count}/{total} passed
- Production build: {status}
- Usage scan: {findings}
```

## Worked Example

See [examples/group-a-babel-traverse.md](examples/group-a-babel-traverse.md) for a complete, validated example that applied this workflow to 6 Babel packages affected by CVE-2023-45133.

## Error Handling

| Scenario | Action |
|----------|--------|
| `npm install` peer dep conflict | Retry with `--legacy-peer-deps` |
| `ng test` pre-existing failure | Document as pre-existing; rely on security tests + build |
| Override breaks build | Revert override, try direct upgrade, or escalate to resolution plan |
| No `gh` CLI available | Output PR body to console for manual creation |
| DevRev MCP unavailable | Use direct `curl` calls to DevRev REST API |

## References

- [Remediation Workflow (detailed)](references/remediation-workflow.md)
- [DevRev Ticket Fetching](references/devrev-ticket-fetching.md)
- [npm Override Patterns](references/npm-override-patterns.md)
- [Security Test Patterns](references/security-test-patterns.md)
- [MCP Configuration](references/mcp-config.md)
- [Example: Babel/traverse](examples/group-a-babel-traverse.md)
