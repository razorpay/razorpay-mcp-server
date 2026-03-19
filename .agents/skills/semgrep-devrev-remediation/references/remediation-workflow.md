# Remediation Workflow (Detailed)

Complete 8-step workflow for remediating SCA-flagged vulnerable dependencies. Execute all steps in order for each dependency group.

---

## Step 1: Extract Remediation Guidance from DevRev Ticket

**Goal:** Get the CVE details, minimum safe version, and remediation steps from the DevRev ticket itself. DevRev SCA tickets already embed all Semgrep finding data — no separate Semgrep MCP server is needed.

### Actions

1. Fetch the DevRev ticket with full body (see [devrev-ticket-fetching.md](devrev-ticket-fetching.md))
2. Extract key fields:
   - `body` — contains CVE description, affected versions, and fix guidance
   - `custom_fields.ctype__issue_url` — Semgrep finding URL (for reference only)
   - `custom_fields.ctype__triggered_rule_name` — Semgrep rule ID
   - `custom_fields.ctype__remediation_guide` — minimum safe version and upgrade path

3. From the ticket body, identify:
   - **CVE ID** (e.g., CVE-2023-45133)
   - **Minimum safe version** (e.g., `>= 7.23.2`)
   - **Vulnerability type** (e.g., arbitrary code execution)

### Example: Extracting from a DevRev Ticket

```bash
curl -X GET "https://api.devrev.ai/works.get?id=${TICKET_DON_ID}" \
  -H "Authorization: Bearer ${DEVREV_PAT}" | \
  python3 -c "
import sys, json
work = json.load(sys.stdin).get('work', {})
print('Title:', work.get('title', ''))
print('Body:', work.get('body', '')[:500])
cf = work.get('custom_fields', {})
print('Remediation:', cf.get('ctype__remediation_guide', 'Not provided'))
print('Semgrep URL:', cf.get('ctype__issue_url', 'N/A'))
"
```

### Decision

| Result | Next Step |
|--------|-----------|
| Remediation guide has safe version | Extract target version, continue to Step 2 |
| Ticket body has CVE but no version | Search web for `{CVE-ID} minimum safe version` |
| No useful data in ticket | Search npm advisory database or GitHub security advisories |

---

## Step 2: Identify Stable Target Versions

**Goal:** Find the latest stable version of the vulnerable package (and related packages) from the npm registry.

### Actions

```bash
# Query npm registry for each affected package
curl -s "https://registry.npmjs.org/{package-name}" | \
  python3 -c "
import sys, json
d = json.load(sys.stdin)
tags = d.get('dist-tags', {})
print(f'Latest: {tags.get(\"latest\", \"?\")}')"

# Check what the latest version depends on (does it still include the vuln dep?)
curl -s "https://registry.npmjs.org/{package-name}/{latest-version}" | \
  python3 -c "
import sys, json
d = json.load(sys.stdin)
deps = d.get('dependencies', {})
print(json.dumps(deps, indent=2))"
```

### Key Checks

- Does the latest version of the parent package still depend on the vulnerable package?
- If the dependency was removed entirely in a newer version, that is the ideal fix
- Cross-reference with the CVE minimum safe version

### Output

A table mapping each package to its current locked version and proposed target:

```
| Package | Current | Target | Strategy |
|---------|---------|--------|----------|
| @babel/traverse | 7.24.7 | ^7.25.0 | npm override |
```

---

## Step 3: Gather Release Notes / Changelogs

**Goal:** Identify breaking changes between the current and target versions.

### Sources (check in order)

1. **npm registry metadata:**
   ```bash
   curl -s "https://registry.npmjs.org/{package}" | \
     python3 -c "import sys,json; d=json.load(sys.stdin); repo=d.get('repository',{}); print(repo.get('url',''))"
   ```

2. **GitHub releases API:**
   ```bash
   curl -s "https://api.github.com/repos/{owner}/{repo}/releases?per_page=20" | \
     python3 -c "
   import sys, json
   for r in json.loads(sys.stdin.read()):
       tag = r['tag_name']
       print(f'{tag}: {r[\"body\"][:200]}')"
   ```

3. **CHANGELOG.md in the repository** (if exists)

### What to Look For

- Lines containing "BREAKING", "breaking change", "removed", "deprecated"
- Major version bumps (semver major = breaking changes expected)
- API surface changes (renamed exports, removed functions)

---

## Step 4: Scan Repo for Usage

**Goal:** Determine if the vulnerable package is used directly in application code, or only as a transitive dependency.

### Actions

```bash
# Search source code for direct imports/requires
grep -r "from '@{package-name}" src/ --include="*.ts" --include="*.js"
grep -r "require('@{package-name}" src/ --include="*.ts" --include="*.js"

# Search config files
grep -r "{package-name}" *.config.* .babelrc .browserslistrc

# Check if it's a direct or dev dependency
grep "{package-name}" package.json
```

### Classification

| Usage Pattern | Risk Level | Fix Approach |
|--------------|------------|--------------|
| No direct imports, transitive only | LOW | Override in `package.json` |
| Imported in source code | MEDIUM | Version bump + code review |
| Used in build config | MEDIUM | Override + build verification |
| Core runtime dependency | HIGH | Careful upgrade + extensive testing |

---

## Step 5: Breaking Change Analysis

**Goal:** Determine if upgrading will break existing functionality.

### Analysis Checklist

1. **Semver analysis:**
   - Patch bump (x.y.Z): No breaking changes expected
   - Minor bump (x.Y.z): New features, backward compatible
   - Major bump (X.y.z): Breaking changes likely

2. **Dependency tree comparison:**
   ```bash
   # Compare dependencies between current and target version
   curl -s "https://registry.npmjs.org/{pkg}/{current}" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin).get('dependencies',{}), indent=2))"
   curl -s "https://registry.npmjs.org/{pkg}/{target}" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin).get('dependencies',{}), indent=2))"
   ```

3. **peerDependencies check:** Ensure the target version's peers are compatible with the project

### Risk Matrix

| Factor | Low Risk | High Risk |
|--------|----------|-----------|
| Semver delta | Patch/minor | Major |
| Direct usage | None (transitive only) | Imported in source |
| Dep type | devDependency | production dependency |
| Test coverage | Good existing tests | No tests |

---

## Step 6: Write Security Verification Tests

**Goal:** Create automated tests that verify the vulnerability is resolved.

### Test File Convention

```
tests/security/{package-name}-{cve-id}.spec.js
```

### Required Assertions

1. **Override exists** in `package.json`
2. **Resolved version** in `package-lock.json` meets the safe threshold
3. **No nested vulnerable copies** exist in the lockfile
4. **Existing overrides preserved** (no regressions)
5. **Lockfile structure is valid**

See [security-test-patterns.md](security-test-patterns.md) for complete test templates.

---

## Step 7: Run Tests

**Goal:** Verify the fix does not introduce regressions.

### Test Execution Order

```bash
# 1. Run security verification tests
npx jasmine tests/security/{test-file}.spec.js

# 2. Run production build
npx ng build --configuration=production   # Angular
# OR
npm run build                              # Generic

# 3. Run existing test suite
npm test
# OR (if Karma missing deps)
npx ng test --watch=false --browsers=ChromeHeadless
```

### Pass/Fail Criteria

| Test | Pass Criteria | Fail Action |
|------|--------------|-------------|
| Security tests | All assertions pass | Debug override; check lockfile |
| Production build | Exit code 0 (warnings OK) | Revert; try alternative version |
| Existing tests | Same pass rate as before | Check if failure is pre-existing |

### Documenting Pre-existing Failures

If `npm test` fails with issues unrelated to the change (e.g., missing `karma-coverage`), document it:

```
Pre-existing test failure: karma-coverage module not found
Not related to @babel/traverse override change.
Evidence: Same failure on master branch without changes.
```

---

## Step 8: Create PR or Resolution Plan

### Step 8a: Tests Pass -> Create PR

```bash
# 1. Create branch
git checkout -b fix/semgrep-{dependency-name}

# 2. Stage only relevant files
git add package.json package-lock.json tests/security/{test-file}.spec.js

# 3. Commit with detailed message
git commit -m "fix(security): add {package} override to resolve {CVE-ID}

{detailed description}

Resolves DevRev tickets: {ISS-IDs}
Semgrep rule: {rule-id}

Verification:
- Security tests: {N}/{N} passed
- Production build: SUCCESS
- Usage scan: {summary}"

# 4. Push and create PR
git push -u origin fix/semgrep-{dependency-name}
gh pr create \
  --base master \
  --head fix/semgrep-{dependency-name} \
  --title "fix(security): resolve {CVE-ID} - {package} override ({priority})" \
  --body "{PR body with summary, root cause, affected packages, verification results}"
```

### Step 8b: Tests Fail -> Resolution Plan

Generate a detailed resolution plan document:

```markdown
# Resolution Plan: {CVE-ID} - {package}

## Status: MANUAL INTERVENTION REQUIRED

## What Failed
- {test name}: {failure reason}
- Build error: {error message}

## Root Cause Analysis
{why the automated fix did not work}

## Recommended Manual Steps
1. {step 1}
2. {step 2}

## Alternative Approaches
- {approach A}: {tradeoffs}
- {approach B}: {tradeoffs}

## DevRev Tickets
{list of affected ticket IDs}
```

Post this as a comment on the DevRev tickets and assign to the development team.

