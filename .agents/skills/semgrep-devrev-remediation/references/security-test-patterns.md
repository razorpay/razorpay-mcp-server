# Security Test Patterns for Dependency Remediation

Templates and patterns for writing security verification tests that confirm vulnerable dependencies have been remediated.

---

## Test File Convention

```
tests/security/{package-name}-{cve-id}.spec.js
```

Example: `tests/security/babel-traverse-cve-2023-45133.spec.js`

## Test Framework

These templates use **Jasmine** (default for Angular projects). Adapt `expect(...).toBeTrue()` to `expect(...).toBe(true)` for Jest.

---

## Template: Override Verification Test

This is the primary test template. It verifies that an npm override is correctly applied and the resolved version is safe.

```javascript
const fs = require('fs');
const path = require('path');

describe('{CVE-ID}: {package-name} override verification', () => {
  let packageJson;
  let lockfile;

  beforeAll(() => {
    packageJson = JSON.parse(
      fs.readFileSync(path.resolve(__dirname, '../../package.json'), 'utf8')
    );
    lockfile = JSON.parse(
      fs.readFileSync(path.resolve(__dirname, '../../package-lock.json'), 'utf8')
    );
  });

  // TEST 1: Override exists in package.json
  it('should have the override in package.json', () => {
    expect(packageJson).toBeDefined();
    expect(packageJson.overrides).toBeDefined();
    expect(packageJson.overrides['{package-name}']).toBeDefined();
    expect(packageJson.overrides['{package-name}']).toEqual('{override-range}');
  });

  // TEST 2: Lockfile is valid
  it('should have a valid lockfile structure', () => {
    expect(lockfile).toBeDefined();
    expect(lockfile.packages).toBeDefined();
  });

  // TEST 3: Resolved version is safe
  it('should resolve {package-name} to a safe version (>= {min-safe-version})', () => {
    const pkg = lockfile.packages['node_modules/{package-name}'];
    expect(pkg).toBeDefined();

    const currentVersion = pkg.version;
    const [major, minor, patch] = currentVersion.split('.').map(Number);

    // Adjust this condition based on the CVE minimum safe version
    const isSafe = (
      major > {SAFE_MAJOR} ||
      (major === {SAFE_MAJOR} && minor > {SAFE_MINOR}) ||
      (major === {SAFE_MAJOR} && minor === {SAFE_MINOR} && patch >= {SAFE_PATCH})
    );

    expect(isSafe).toBeTrue();
    console.log(`  {package-name}: ${currentVersion} (safe: >= {min-safe-version})`);
  });

  // TEST 4+: Parent packages still resolve correctly
  it('should ensure {parent-package} resolves correctly', () => {
    const pkg = lockfile.packages['node_modules/{parent-package}'];
    expect(pkg).toBeDefined();
    expect(pkg.version).toEqual('{expected-version}');
    console.log(`  {parent-package}: ${pkg.version}`);
  });
});
```

### Placeholder Reference

| Placeholder | Replace With | Example |
|-------------|-------------|---------|
| `{CVE-ID}` | CVE identifier | `CVE-2023-45133` |
| `{package-name}` | Vulnerable npm package | `@babel/traverse` |
| `{override-range}` | Range in overrides | `^7.25.0` |
| `{min-safe-version}` | Minimum safe version | `7.23.2` |
| `{SAFE_MAJOR}` | Major part of min safe | `7` |
| `{SAFE_MINOR}` | Minor part of min safe | `23` |
| `{SAFE_PATCH}` | Patch part of min safe | `2` |
| `{parent-package}` | Package that depends on vuln pkg | `babel-plugin-polyfill-regenerator` |
| `{expected-version}` | Expected locked version of parent | `0.4.1` |

---

## Template: No-Nested-Copies Test

Verifies that no nested copies of the vulnerable package exist at unsafe versions.

```javascript
it('should not have nested vulnerable copies of {package-name}', () => {
  const vulnerablePaths = [];

  for (const [pkgPath, pkgInfo] of Object.entries(lockfile.packages || {})) {
    if (pkgPath.includes('{package-name}') && pkgInfo.version) {
      const [major, minor, patch] = pkgInfo.version.split('.').map(Number);
      const isSafe = (
        major > {SAFE_MAJOR} ||
        (major === {SAFE_MAJOR} && minor > {SAFE_MINOR}) ||
        (major === {SAFE_MAJOR} && minor === {SAFE_MINOR} && patch >= {SAFE_PATCH})
      );
      if (!isSafe) {
        vulnerablePaths.push({ path: pkgPath, version: pkgInfo.version });
      }
    }
  }

  if (vulnerablePaths.length > 0) {
    console.log('  Vulnerable copies found:');
    vulnerablePaths.forEach(v => console.log(`    ${v.path}: ${v.version}`));
  }

  expect(vulnerablePaths.length).toBe(0);
});
```

---

## Template: Existing Overrides Preserved

Ensures that adding a new override did not accidentally remove existing ones.

```javascript
it('should preserve all existing overrides', () => {
  const expectedOverrides = {
    'http-proxy-middleware': '^2.0.9',
    '@babel/helpers': '^7.26.10',
    '@babel/runtime': '^7.26.10',
    'esbuild': '^0.25.0',
    // NEW: the override being added
    '{package-name}': '{override-range}'
  };

  for (const [pkg, range] of Object.entries(expectedOverrides)) {
    expect(packageJson.overrides[pkg]).toEqual(range);
  }
});
```

---

## Running Security Tests

### Jasmine (Angular Projects)

```bash
# Install jasmine if not present
npm install --save-dev jasmine

# Run specific security test
npx jasmine tests/security/{test-file}.spec.js
```

### Jest (React/Node Projects)

```bash
# Run specific security test
npx jest tests/security/{test-file}.spec.js
```

### Standalone Node.js (No Framework)

If no test framework is available, use Node.js assertions:

```javascript
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
const lockfile = JSON.parse(fs.readFileSync('package-lock.json', 'utf8'));

// Test 1: Override exists
assert.ok(packageJson.overrides, 'overrides section missing');
assert.ok(packageJson.overrides['{package-name}'], 'override missing for {package-name}');

// Test 2: Safe version resolved
const pkg = lockfile.packages['node_modules/{package-name}'];
assert.ok(pkg, '{package-name} not found in lockfile');
const [major, minor, patch] = pkg.version.split('.').map(Number);
assert.ok(
  major > {SAFE_MAJOR} || (major === {SAFE_MAJOR} && minor >= {SAFE_MINOR}),
  `Unsafe version: ${pkg.version}`
);

console.log('All security checks passed');
```

Run with: `node tests/security/{test-file}.js`

---

## Build Verification

In addition to lockfile tests, verify the production build still succeeds:

```bash
# Angular
npx ng build --configuration=production 2>&1 | tail -5

# React (Create React App)
npm run build 2>&1 | tail -5

# Generic
npm run build-prod 2>&1 | tail -5
```

A successful build with the override confirms the new version is compatible with the build toolchain.

---

## Test Output Interpretation

### All Passing

```
6 specs, 0 failures
Finished in 0.05 seconds

  @babel/traverse: 7.29.0 (safe: >= 7.23.2)
  babel-plugin-polyfill-regenerator: 0.4.1
  babel-plugin-polyfill-corejs2: 0.3.3
  ...
```

Proceed to Step 8a (Create PR).

### Failures

```
3 specs, 1 failure

  should resolve @babel/traverse to a safe version
    Expected false to be true.
```

This means the override did not take effect. Debug steps:

1. Verify `package.json` overrides section is syntactically correct
2. Delete `node_modules` and `package-lock.json`, then `npm install`
3. Check if a nested lockfile or workspace config is overriding the top-level override
4. If the override range does not match any published version, adjust the range

