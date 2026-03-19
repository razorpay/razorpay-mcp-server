# API Contract Validation Checks

## Overview

Validates API endpoint contracts including request/response schemas, error handling, timeout/retry configurations, and backward compatibility.

**Load when:** PR modifies API handlers, request/response models, or API documentation

**Total Checks:** 4

**Severity Distribution:**
- 🚨 Critical: 2
- ⚠️ High: 1
- 📋 Medium: 1

---

## Check 1: Request Schema Validation 🚨 CRITICAL

### What to Check

API endpoints must validate request payloads before processing.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No validation
func CreateTerminal(c *gin.Context) {
    var req TerminalRequest

    // ❌ No validation - binds and processes directly
    c.BindJSON(&req)

    // ❌ Could have missing/invalid fields
    terminal := &Terminal{
        ID:         utils.NewID(),
        MerchantID: req.MerchantID,  // Could be empty!
        Gateway:    req.Gateway,      // Could be invalid!
    }

    repo.Save(terminal)
    c.JSON(200, terminal)
}
```

**Problem:**
- Invalid data reaches business logic
- Database constraint violations
- Poor error messages to clients

### Good Pattern ✅

```go
// CORRECT: Request validation with tags
type TerminalRequest struct {
    MerchantID string `json:"merchant_id" binding:"required,len=14" validate:"rzp_id"`
    Gateway    string `json:"gateway" binding:"required,oneof=hitachi cybersource"`
    Status     string `json:"status" binding:"required,oneof=active inactive"`
    OrgID      string `json:"org_id" binding:"required,len=14"`
}

func CreateTerminal(c *gin.Context) {
    var req TerminalRequest

    // ✅ Validate on bind
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn(c, "request_validation_failed", "error", err)
        c.JSON(400, gin.H{
            "error":   "validation_failed",
            "details": formatValidationErrors(err),
        })
        return
    }

    // ✅ Additional business validation
    if err := validateTerminalRequest(c, req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    terminal := buildTerminal(req)

    if err := repo.Save(terminal); err != nil {
        logger.Error(c, "terminal_save_failed", "error", err)
        c.JSON(500, gin.H{"error": "failed to create terminal"})
        return
    }

    c.JSON(201, terminal)
}

// Business validation
func validateTerminalRequest(ctx *gin.Context, req TerminalRequest) error {
    // ✅ Check merchant exists
    merchant, err := repo.FindMerchant(req.MerchantID)
    if err != nil || merchant == nil {
        return errors.New("merchant not found")
    }

    // ✅ Check org belongs to merchant
    if merchant.OrgID != req.OrgID {
        return errors.New("org does not belong to merchant")
    }

    // ✅ Check gateway supported
    if !isGatewaySupported(req.Gateway) {
        return fmt.Errorf("gateway %s not supported", req.Gateway)
    }

    return nil
}
```

### Detection Strategy

```bash
# Find API handlers
grep -n "func.*gin.Context" internal/handlers/*.go

# For each handler:
# 1. Check request struct has validation tags
# 2. Verify ShouldBindJSON() error is checked
# 3. Check 400 response returned on validation error
```

### Flag Conditions

Flag if:
- Request struct without validation tags (`binding`, `validate`)
- `BindJSON()` error not checked
- No validation error response
- POST/PUT/PATCH endpoint without request validation

### Severity

🚨 **Critical** - Invalid data in system, poor API contract

### Reference

Based on terminals API patterns

---

## Check 2: Response Schema Consistency 🚨 CRITICAL

### What to Check

API responses must follow consistent schema and include proper error formats.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Inconsistent responses
func GetTerminal(c *gin.Context) {
    id := c.Param("id")

    terminal, err := repo.FindTerminal(id)
    if err != nil {
        // ❌ Inconsistent error format
        c.JSON(500, "error finding terminal")
        return
    }

    if terminal == nil {
        // ❌ Different error format
        c.JSON(404, gin.H{"message": "not found"})
        return
    }

    // ❌ No consistent wrapper
    c.JSON(200, terminal)
}

func ListTerminals(c *gin.Context) {
    terminals, _ := repo.ListTerminals()

    // ❌ Different format than GetTerminal
    c.JSON(200, gin.H{
        "data": terminals,
        "count": len(terminals),
    })
}
```

**Problem:**
- Clients can't parse responses consistently
- Different error formats
- Breaking changes to clients

### Good Pattern ✅

```go
// CORRECT: Consistent response schema

// Standard response wrapper
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// Success response helper
func Success(c *gin.Context, code int, data interface{}) {
    c.JSON(code, APIResponse{
        Success: true,
        Data:    data,
    })
}

// Error response helper
func ErrorResponse(c *gin.Context, code int, errorCode, message string, details interface{}) {
    c.JSON(code, APIResponse{
        Success: false,
        Error: &APIError{
            Code:    errorCode,
            Message: message,
            Details: details,
        },
    })
}

// ✅ Consistent usage
func GetTerminal(c *gin.Context) {
    id := c.Param("id")

    terminal, err := repo.FindTerminal(id)
    if err != nil {
        logger.Error(c, "terminal_find_failed", "error", err)
        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to retrieve terminal", nil)
        return
    }

    if terminal == nil {
        ErrorResponse(c, 404, "TERMINAL_NOT_FOUND", "Terminal not found", nil)
        return
    }

    Success(c, 200, terminal)
}

func ListTerminals(c *gin.Context) {
    terminals, err := repo.ListTerminals()
    if err != nil {
        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to list terminals", nil)
        return
    }

    // ✅ Consistent format with GetTerminal
    Success(c, 200, gin.H{
        "terminals": terminals,
        "count":     len(terminals),
    })
}
```

### Severity

🚨 **Critical** - Poor API contract, client integration issues

---

## Check 3: HTTP Status Code Correctness ⚠️ HIGH

### What to Check

API endpoints must use correct HTTP status codes for different scenarios.

### Bad Pattern ❌

```go
// ANTI-PATTERN: Incorrect status codes
func CreateTerminal(c *gin.Context) {
    var req TerminalRequest

    if err := c.BindJSON(&req); err != nil {
        // ❌ 500 for validation error (should be 400)
        c.JSON(500, gin.H{"error": "invalid request"})
        return
    }

    terminal, err := service.CreateTerminal(req)
    if err != nil {
        if err == ErrTerminalAlreadyExists {
            // ❌ 400 for conflict (should be 409)
            c.JSON(400, gin.H{"error": "already exists"})
            return
        }

        // ❌ Generic 500 for all errors
        c.JSON(500, gin.H{"error": "error"})
        return
    }

    // ❌ 200 for creation (should be 201)
    c.JSON(200, terminal)
}

func DeleteTerminal(c *gin.Context) {
    err := service.DeleteTerminal(id)
    if err != nil {
        // ❌ 500 for not found (should be 404)
        c.JSON(500, gin.H{"error": "error deleting"})
        return
    }

    // ❌ 200 with empty body (should be 204)
    c.JSON(200, gin.H{})
}
```

**Problem:**
- Confusing error semantics
- Clients can't distinguish error types
- Poor API design

### Good Pattern ✅

```go
// CORRECT: Proper status codes

func CreateTerminal(c *gin.Context) {
    var req TerminalRequest

    if err := c.BindJSON(&req); err != nil {
        // ✅ 400 Bad Request for validation errors
        ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request", err.Error())
        return
    }

    terminal, err := service.CreateTerminal(req)
    if err != nil {
        if errors.Is(err, ErrTerminalAlreadyExists) {
            // ✅ 409 Conflict for duplicate
            ErrorResponse(c, 409, "TERMINAL_EXISTS", "Terminal already exists", nil)
            return
        }

        if errors.Is(err, ErrMerchantNotFound) {
            // ✅ 404 Not Found for missing dependency
            ErrorResponse(c, 404, "MERCHANT_NOT_FOUND", "Merchant not found", nil)
            return
        }

        if errors.Is(err, ErrUnauthorized) {
            // ✅ 403 Forbidden for authorization issues
            ErrorResponse(c, 403, "FORBIDDEN", "Insufficient permissions", nil)
            return
        }

        // ✅ 500 Internal Error for unexpected issues
        logger.Error(c, "terminal_create_failed", "error", err)
        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to create terminal", nil)
        return
    }

    // ✅ 201 Created for successful creation
    Success(c, 201, terminal)
}

func GetTerminal(c *gin.Context) {
    id := c.Param("id")

    terminal, err := service.GetTerminal(id)
    if err != nil {
        if errors.Is(err, ErrTerminalNotFound) {
            // ✅ 404 Not Found
            ErrorResponse(c, 404, "TERMINAL_NOT_FOUND", "Terminal not found", nil)
            return
        }

        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to retrieve terminal", nil)
        return
    }

    // ✅ 200 OK for successful retrieval
    Success(c, 200, terminal)
}

func UpdateTerminal(c *gin.Context) {
    // ... update logic

    // ✅ 200 OK with updated resource
    Success(c, 200, terminal)
}

func DeleteTerminal(c *gin.Context) {
    err := service.DeleteTerminal(id)
    if err != nil {
        if errors.Is(err, ErrTerminalNotFound) {
            ErrorResponse(c, 404, "TERMINAL_NOT_FOUND", "Terminal not found", nil)
            return
        }

        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to delete terminal", nil)
        return
    }

    // ✅ 204 No Content for successful deletion
    c.Status(204)
}
```

**Status Code Guidelines:**
- **200 OK**: Successful GET, PUT, PATCH
- **201 Created**: Successful POST (creation)
- **204 No Content**: Successful DELETE
- **400 Bad Request**: Validation errors
- **401 Unauthorized**: Missing/invalid auth
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource already exists
- **500 Internal Server Error**: Unexpected errors

### Severity

⚠️ **High** - Poor API semantics, client confusion

---

## Check 4: Timeout and Retry Headers 📋 MEDIUM

### What to Check

APIs that **introduce new timeout middleware or explicitly document retry semantics** should set appropriate response headers. This is NOT a universal requirement — `X-Retryable` and `Retry-After` are non-standard headers not used by most Razorpay internal services, so flagging all APIs without them would produce near-universal false positives.

**Only flag if:** The PR adds a new timeout middleware, a new long-running endpoint, or explicitly claims retry-safety in comments/docs.

### Bad Pattern ❌

```go
// ANTI-PATTERN: No timeout guidance
func ProcessPayment(c *gin.Context) {
    // ❌ Long-running operation with no timeout indication
    result := gateway.ProcessPayment(payment)  // Can take 30+ seconds

    c.JSON(200, result)  // ❌ Client doesn't know to wait
}

// ANTI-PATTERN: No retry guidance
func CreateTerminal(c *gin.Context) {
    err := service.CreateTerminal(req)
    if err != nil {
        // ❌ No indication if client should retry
        c.JSON(500, gin.H{"error": "creation failed"})
        return
    }
}
```

### Good Pattern ✅

```go
// CORRECT: Add timeout and retry headers

func ProcessPayment(c *gin.Context) {
    // ✅ Set request timeout header
    c.Header("X-Request-Timeout", "30s")

    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()

    result, err := gateway.ProcessPaymentWithContext(ctx, payment)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            // ✅ Indicate retry-ability
            c.Header("Retry-After", "60")  // Retry after 60 seconds
            ErrorResponse(c, 504, "TIMEOUT", "Payment processing timeout", nil)
            return
        }

        // ✅ Indicate non-retryable error
        c.Header("X-Retryable", "false")
        ErrorResponse(c, 400, "PAYMENT_FAILED", "Payment declined", err.Error())
        return
    }

    c.JSON(200, result)
}

// CORRECT: Document retry semantics
func CreateTerminal(c *gin.Context) {
    terminal, err := service.CreateTerminal(req)
    if err != nil {
        if errors.Is(err, ErrTerminalAlreadyExists) {
            // ✅ Conflict - safe to retry with different data
            c.Header("X-Retryable", "false")  // Don't retry same request
            ErrorResponse(c, 409, "TERMINAL_EXISTS", "Terminal already exists", nil)
            return
        }

        // ✅ Internal error - safe to retry
        c.Header("X-Retryable", "true")
        c.Header("Retry-After", "5")
        ErrorResponse(c, 500, "INTERNAL_ERROR", "Failed to create terminal", nil)
        return
    }

    c.JSON(201, terminal)
}

// CORRECT: Add timeout middleware
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)

        // ✅ Set timeout header
        c.Header("X-Request-Timeout", timeout.String())

        done := make(chan struct{})
        go func() {
            c.Next()
            close(done)
        }()

        select {
        case <-done:
            // Request completed
        case <-ctx.Done():
            // ✅ Timeout occurred
            c.Header("Retry-After", "10")
            ErrorResponse(c, 504, "REQUEST_TIMEOUT", "Request timeout", nil)
            c.Abort()
        }
    }
}
```

### Detection Strategy

```bash
# Only check if the PR introduces timeout middleware or long-running operation patterns
git diff main..HEAD | grep -qE "TimeoutMiddleware|WithTimeout|context\.WithTimeout" || exit 0

# If timeout context is added, verify timeout behavior is surfaced to clients
grep -n "context\.WithTimeout" <pr_files> | while read line; do
    line_num=$(echo "$line" | cut -d: -f1)
    file=$(echo "$line" | cut -d: -f1)
    # Check if there's a 504 or timeout error response nearby
    context=$(sed -n "$((line_num)),$((line_num+20))p" "$file")
    if ! echo "$context" | grep -qE "504|TimeoutError|REQUEST_TIMEOUT|Retry-After"; then
        echo "📋 $file:$line_num: Timeout context set but no timeout error response — clients won't know to retry"
    fi
done
```

### Severity

📋 **Medium** - Applicable only when timeout/retry patterns are explicitly introduced; not a universal API requirement

---

## Summary Table

| Check # | Pattern | Severity | Risk |
|---------|---------|----------|------|
| 1 | Request validation | 🚨 Critical | Invalid data in system |
| 2 | Response consistency | 🚨 Critical | Client integration issues |
| 3 | Status code correctness | ⚠️ High | API semantics confusion |
| 4 | Timeout/retry headers (when added) | 📋 Medium | Client timeout transparency |

---

## How to Apply

**For each file matching** `internal/handlers/*`, `internal/api/*`:

1. Check request structs have validation tags
2. Verify response format is consistent
3. Check HTTP status codes are correct
4. Look for timeout and retry headers

**Example output:**

```
📁 File: internal/handlers/terminal_handler.go

🚨 Check #1 Failed: No request validation (Line 23)
   Code: c.BindJSON(&req)
   Fix: Add validation tags and check error

🚨 Check #2 Failed: Inconsistent response format (Line 45)
   Code: c.JSON(200, terminal)
   Fix: Use Success() helper for consistent wrapper

⚠️  Check #3 Failed: Wrong status code (Line 67)
   Code: c.JSON(500, ...) for validation error
   Fix: Use 400 for validation errors

✅ Check #4 Passed: Timeout headers present
```
