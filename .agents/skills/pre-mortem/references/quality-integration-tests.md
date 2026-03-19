# Integration Test (SLIT) Quality Checks

## Overview

Validates Service-Level Integration Tests (SLITs) exist for API endpoints and cover error scenarios.

**Load when:** PR adds/modifies API endpoints in `internal/handlers/*` or `internal/api/*`

**Total Checks:** 2

**Severity Distribution:**
- ⚠️ High: 2

---

## Check 1: SLIT Exists for New Endpoints ⚠️ HIGH

### What to Check

New API endpoints must have corresponding SLIT (Service-Level Integration Test).

### Bad Pattern ❌

```
PR adds:
✅ internal/handlers/terminal_handler.go
    - POST /terminals
    - GET /terminals/:id
    - PUT /terminals/:id

❌ No SLITs in test/slit/ directory
```

**Problem:**
- API not tested end-to-end
- Database integration not validated
- No validation of request/response flow

### Good Pattern ✅

```
PR includes:
✅ internal/handlers/terminal_handler.go
    - POST /terminals
    - GET /terminals/:id
    - PUT /terminals/:id

✅ test/slit/terminal_slit_test.go
    - TestCreateTerminal_Success
    - TestGetTerminal_Success
    - TestGetTerminal_NotFound
    - TestUpdateTerminal_Success
```

**SLIT Example:**

```go
// test/slit/terminal_slit_test.go

func TestCreateTerminal_Success(t *testing.T) {
    // ✅ Full integration test
    // - Hits actual HTTP endpoint
    // - Writes to real test database
    // - Validates full request/response cycle

    req := httptest.NewRequest("POST", "/terminals", strings.NewReader(`{
        "merchant_id": "merch_123456789",
        "gateway": "hitachi",
        "status": "active"
    }`))
    req.Header.Set("Content-Type", "application/json")

    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    // Validate response
    assert.Equal(t, 201, resp.Code)

    var terminal Terminal
    json.Unmarshal(resp.Body.Bytes(), &terminal)
    assert.NotEmpty(t, terminal.ID)
    assert.Equal(t, "merch_123456789", terminal.MerchantID)

    // Validate database record created
    dbTerminal, err := repo.FindTerminal(terminal.ID)
    assert.NoError(t, err)
    assert.Equal(t, "active", dbTerminal.Status)
}

func TestGetTerminal_NotFound(t *testing.T) {
    // ✅ Tests error scenario
    req := httptest.NewRequest("GET", "/terminals/invalid_id", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 404, resp.Code)

    var errResp ErrorResponse
    json.Unmarshal(resp.Body.Bytes(), &errResp)
    assert.Equal(t, "TERMINAL_NOT_FOUND", errResp.Error.Code)
}
```

### Detection Strategy

```bash
# Find new/modified API handler files
HANDLER_FILES=$(git diff --name-only main..HEAD | grep "internal/handlers\|internal/api" | grep -v "_test\.go$")

# For each handler file
for file in $HANDLER_FILES; do
    # Extract endpoint functions (Create, Get, Update, Delete, List)
    endpoints=$(grep -E "func.*\(c \*gin\.Context\)" "$file" | awk '{print $2}' | cut -d'(' -f1)

    # Check if SLIT exists
    slit_file="test/slit/$(basename $file _handler.go)_slit_test.go"

    if [ ! -f "$slit_file" ]; then
        FLAG: "SLIT missing for $file"
    else
        # Check if endpoints are tested
        for endpoint in $endpoints; do
            if ! grep -q "Test$endpoint" "$slit_file"; then
                FLAG: "SLIT missing for endpoint $endpoint"
            fi
        done
    fi
done
```

### Flag Conditions

Flag if:
- New API endpoint added
- No SLIT file exists
- SLIT doesn't cover the new endpoint
- Modification to existing endpoint without updating SLIT

### Severity

⚠️ **High** - API not validated end-to-end

### Reference

Based on terminals SLIT patterns in `test/slit/`

---

## Check 2: SLIT Covers Error Scenarios ⚠️ HIGH

### What to Check

SLITs must test error cases, not just happy path.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Only happy path tested
func TestCreateTerminal(t *testing.T) {
    // ✅ Happy path
    req := makeValidRequest()
    resp := callEndpoint(req)

    assert.Equal(t, 201, resp.Code)

    // ❌ Missing error scenarios:
    // - Invalid request (400)
    // - Duplicate terminal (409)
    // - Merchant not found (404)
    // - Database error (500)
    // - Unauthorized (401)
}
```

**Problem:**
- Error handling not validated
- Edge cases not covered
- Production errors not caught

### Good Pattern ✅

```go
// CORRECT: Comprehensive SLIT with error scenarios

func TestCreateTerminal_ValidationError(t *testing.T) {
    // ✅ Test 400 Bad Request
    req := httptest.NewRequest("POST", "/terminals", strings.NewReader(`{
        "merchant_id": "",
        "gateway": "invalid"
    }`))

    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 400, resp.Code)

    var errResp ErrorResponse
    json.Unmarshal(resp.Body.Bytes(), &errResp)
    assert.Equal(t, "VALIDATION_ERROR", errResp.Error.Code)
    assert.Contains(t, errResp.Error.Message, "merchant_id")
}

func TestCreateTerminal_DuplicateConflict(t *testing.T) {
    // ✅ Test 409 Conflict
    // Create terminal first
    createTerminal(t, validRequest)

    // Try to create duplicate
    req := httptest.NewRequest("POST", "/terminals", strings.NewReader(validRequest))
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 409, resp.Code)

    var errResp ErrorResponse
    json.Unmarshal(resp.Body.Bytes(), &errResp)
    assert.Equal(t, "TERMINAL_EXISTS", errResp.Error.Code)
}

func TestCreateTerminal_MerchantNotFound(t *testing.T) {
    // ✅ Test 404 Not Found
    req := httptest.NewRequest("POST", "/terminals", strings.NewReader(`{
        "merchant_id": "nonexistent_merchant",
        "gateway": "hitachi"
    }`))

    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 404, resp.Code)

    var errResp ErrorResponse
    json.Unmarshal(resp.Body.Bytes(), &errResp)
    assert.Equal(t, "MERCHANT_NOT_FOUND", errResp.Error.Code)
}

func TestGetTerminal_NotFound(t *testing.T) {
    // ✅ Test 404 for GET
    req := httptest.NewRequest("GET", "/terminals/nonexistent_id", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 404, resp.Code)
}

func TestUpdateTerminal_Unauthorized(t *testing.T) {
    // ✅ Test 401 Unauthorized
    req := httptest.NewRequest("PUT", "/terminals/term_123", strings.NewReader(`{...}`))
    // No auth header
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    assert.Equal(t, 401, resp.Code)
}
```

### Recommended Error Scenarios per Endpoint

**POST /resource (Create):**
- ✅ 201 Created (happy path)
- ✅ 400 Bad Request (validation error)
- ✅ 404 Not Found (parent resource missing)
- ✅ 409 Conflict (duplicate)
- ✅ 500 Internal Error (database failure)

**GET /resource/:id (Retrieve):**
- ✅ 200 OK (happy path)
- ✅ 404 Not Found
- ✅ 401 Unauthorized (if auth required)

**PUT /resource/:id (Update):**
- ✅ 200 OK (happy path)
- ✅ 400 Bad Request (validation)
- ✅ 404 Not Found
- ✅ 409 Conflict (version conflict)

**DELETE /resource/:id (Delete):**
- ✅ 204 No Content (happy path)
- ✅ 404 Not Found
- ✅ 409 Conflict (resource in use)

### Detection Strategy

```bash
# For each SLIT file
for slit in test/slit/*_test.go; do
    # Count test functions
    total_tests=$(grep -c "^func Test" "$slit")

    # Count error scenario tests
    error_tests=$(grep -c "Error\|NotFound\|Conflict\|Unauthorized\|BadRequest" "$slit")

    # Calculate ratio
    error_ratio=$(echo "scale=2; $error_tests / $total_tests" | bc)

    # Flag if < 40% of tests are error scenarios
    if (( $(echo "$error_ratio < 0.4" | bc -l) )); then
        FLAG: "Insufficient error scenario coverage in $slit"
    fi
done
```

### Flag Conditions

Flag if:
- Less than 40% of SLIT tests are error scenarios
- No 404 test for GET/PUT/DELETE endpoints
- No 400 test for POST/PUT endpoints
- No 409 test for POST (creation) endpoints

### Severity

⚠️ **High** - Error handling not validated

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | SLIT exists for endpoints | ⚠️ High | API not validated |
| 2 | Error scenarios covered | ⚠️ High | Error handling untested |

---

## How to Apply

**For each PR:**

1. Find new/modified API handlers
2. Check SLIT files exist
3. Verify endpoints are tested
4. Check error scenario coverage

**Example output:**

```
📁 Endpoint: POST /terminals

⚠️  Check #1 Failed: SLIT missing
   Handler: internal/handlers/terminal_handler.go:CreateTerminal
   Missing: test/slit/terminal_slit_test.go

   Recommendation: Add SLIT with test cases:
     - TestCreateTerminal_Success (201)
     - TestCreateTerminal_ValidationError (400)
     - TestCreateTerminal_DuplicateConflict (409)
     - TestCreateTerminal_MerchantNotFound (404)

📁 File: test/slit/payment_slit_test.go

⚠️  Check #2 Failed: Missing error scenarios
   Total tests: 5
   Error scenario tests: 1 (20%)
   Expected: ≥40%

   Missing error tests:
     - TestProcessPayment_InsufficientFunds (400)
     - TestProcessPayment_GatewayTimeout (504)
     - TestProcessPayment_DuplicatePayment (409)

✅ GET /terminals/:id: SLIT complete with error scenarios
```

---

## SLIT vs Unit Test

**SLIT (Service-Level Integration Test):**
- Tests full HTTP request/response cycle
- Uses real database (test instance)
- Validates routing, middleware, handler, repository
- Slower but more comprehensive

**Unit Test:**
- Tests individual functions in isolation
- Uses mocks for database
- Fast, focused on logic
- More granular

**Both are required:**
- Unit tests: 80%+ coverage of business logic
- SLITs: 100% coverage of API endpoints
