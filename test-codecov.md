# Codecov Status Check Test

This is a test file to trigger CI and verify Codecov status checks work.

## Expected Status Checks:
- ✅ Run tests and publish test coverage (CI)
- ❌ codecov/project (Overall coverage ≥ 75%) - Should FAIL
- ❌ codecov/patch (New code coverage ≥ 75%) - Should FAIL

Current coverage: ~65.9%
Expected result: ❌ codecov/project should FAIL (below 75%)

Testing CI trigger...
