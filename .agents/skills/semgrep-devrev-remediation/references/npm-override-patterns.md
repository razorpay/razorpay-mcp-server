# npm Override Patterns for Transitive Dependency Vulnerabilities

How to use `npm overrides` to fix vulnerable transitive dependencies without upgrading direct dependencies.

---

## When to Use Overrides

| Scenario | Use Override? | Alternative |
|----------|--------------|-------------|
| Vuln package is a transitive dep of a devDependency | YES | Wait for parent package to upgrade |
| Vuln package is a direct dependency | NO | Bump version directly in `dependencies` |
| Parent package pins an exact vulnerable version | YES | Fork parent package (last resort) |
| Multiple parents depend on the same vuln package | YES (single override fixes all) | Upgrade each parent individually |

### Key Advantage

A single override entry forces **all** copies of a package (at any depth in the tree) to resolve to the specified version range. This avoids waiting for every intermediate parent to release a patch.

## How npm Overrides Work

The `overrides` field in `package.json` (npm 8.3+) replaces the resolved version of any matching package in the dependency tree.

```json
{
  "overrides": {
    "@babel/traverse": "^7.25.0"
  }
}
```

This tells npm: "Wherever any package depends on `@babel/traverse`, resolve it to a version matching `^7.25.0` instead of whatever range the parent requested."

### Syntax Options

```json
{
  "overrides": {
    // Force all instances to match this range
    "vulnerable-pkg": "^SAFE_VERSION",

    // Override only when required by a specific parent
    "parent-pkg": {
      "vulnerable-pkg": "^SAFE_VERSION"
    },

    // Use the version from the top-level dependencies
    "vulnerable-pkg": "$vulnerable-pkg"
  }
}
```

## Step-by-Step: Adding an Override

### 1. Identify the Current Vulnerable Version

```bash
# Check what version is currently resolved
cat package-lock.json | python3 -c "
import sys, json
lock = json.load(sys.stdin)
pkg = lock.get('packages', {}).get('node_modules/@babel/traverse', {})
print(f'Resolved: {pkg.get(\"version\", \"not found\")}')
"
```

### 2. Determine the Safe Version

From the CVE advisory or Semgrep finding, identify the minimum safe version:

```
CVE-2023-45133: @babel/traverse >= 7.23.2 is safe
```

Choose the latest minor/patch within the safe range for maximum compatibility:

```bash
# Check latest available version
curl -s "https://registry.npmjs.org/@babel/traverse" | \
  python3 -c "import sys,json; print(json.load(sys.stdin)['dist-tags']['latest'])"
```

### 3. Add the Override to package.json

Add or extend the `overrides` section:

```json
{
  "name": "your-project",
  "dependencies": { ... },
  "devDependencies": { ... },
  "overrides": {
    "@babel/traverse": "^7.25.0"
  }
}
```

### 4. Run npm install

```bash
# Clean install to apply overrides
rm -rf node_modules
npm install

# If peer dependency conflicts occur:
npm install --legacy-peer-deps
```

### 5. Verify the Override Took Effect

```bash
# Check the resolved version in package-lock.json
cat package-lock.json | python3 -c "
import sys, json
lock = json.load(sys.stdin)
pkg = lock.get('packages', {}).get('node_modules/@babel/traverse', {})
version = pkg.get('version', 'NOT FOUND')
print(f'Resolved version: {version}')
# Verify it meets the minimum safe version
major, minor, patch = map(int, version.split('.'))
safe = major > 7 or (major == 7 and minor >= 25)
print(f'Safe: {safe}')
"
```

### 6. Check for Nested Copies

Some packages may install their own nested copy. Verify no vulnerable copies remain:

```bash
# Search for all instances of the package in the lockfile
cat package-lock.json | python3 -c "
import sys, json
lock = json.load(sys.stdin)
for path, info in lock.get('packages', {}).items():
    if '@babel/traverse' in path:
        print(f'{path}: {info.get(\"version\", \"?\")}')"
```

If nested copies exist at vulnerable versions, the override scope may need adjustment.

## Common Patterns

### Pattern A: Build-Tool Transitive Dependency

The most common pattern for Angular/React projects. Build tools pull in vulnerable packages as transitive dependencies, but the packages are never used in production runtime code.

```
@angular-devkit/build-angular (devDependency)
  └── @babel/preset-env
       └── babel-plugin-polyfill-regenerator
            └── @babel/helper-define-polyfill-provider
                 └── @babel/traverse@7.24.7  <-- VULNERABLE
```

**Fix:**

```json
{
  "overrides": {
    "@babel/traverse": "^7.25.0"
  }
}
```

**Risk:** LOW - build tools only run at build time, not in production. A minor/patch override is safe.

### Pattern B: Multiple Overrides for Related Packages

When a CVE affects multiple packages in the same ecosystem:

```json
{
  "overrides": {
    "http-proxy-middleware": "^2.0.9",
    "@babel/helpers": "^7.26.10",
    "@babel/runtime": "^7.26.10",
    "esbuild": "^0.25.0",
    "@babel/traverse": "^7.25.0"
  }
}
```

### Pattern C: Scoped Override (Parent-Specific)

When only one parent needs the override (to avoid affecting other consumers):

```json
{
  "overrides": {
    "babel-plugin-polyfill-regenerator": {
      "@babel/traverse": "^7.25.0"
    }
  }
}
```

## Verification Checklist

After applying any override:

- [ ] `npm install` completes without errors
- [ ] `package-lock.json` shows the safe version for the overridden package
- [ ] No nested copies remain at vulnerable versions
- [ ] Production build succeeds (`npm run build` or `ng build --configuration=production`)
- [ ] Existing tests pass (or failures are pre-existing and documented)
- [ ] Security verification tests pass (see [security-test-patterns.md](security-test-patterns.md))

## Gotchas

1. **npm overrides require npm 8.3+.** Check with `npm --version`. Yarn uses `resolutions` instead.
2. **Overrides apply globally** by default. Use scoped overrides if the package is also a direct dependency.
3. **`package-lock.json` must be regenerated** after adding overrides. Always run `npm install` after modifying overrides.
4. **Peer dependency conflicts** may occur. Use `--legacy-peer-deps` if the conflict is pre-existing and unrelated to the override.
5. **Overrides do not appear in `npm ls` output** by default. Always verify via `package-lock.json` directly.

