# Security Assessment Checklist for Go Code Review

## Overview

Security vulnerability assessment framework for Layer 3 quality review. Provides systematic security risk evaluation and rating.

---

## Security Risk Rating Scale

| Level | Description | Examples |
|-------|-------------|----------|
| **None** | No security concerns identified | Well-validated inputs, proper auth |
| **Low** | Minor issues, unlikely to be exploited | Missing validation on non-critical field |
| **Medium** | Exploitable under specific conditions | Partial input validation, weak crypto config |
| **High** | Likely exploitable, significant impact | SQL injection, auth bypass, exposed secrets |
| **Critical** | Easily exploitable, severe impact | RCE, privilege escalation, data corruption |

---

## Security Assessment Categories

### 1. Input Validation

**Checklist**:
- [ ] All external inputs validated before use
- [ ] Length limits enforced on string inputs
- [ ] Numeric inputs checked for range/overflow
- [ ] Regex patterns validated (no ReDoS vulnerability)
- [ ] File uploads validated (type, size, content)

**Common Vulnerabilities**:
```go
// ❌ HIGH RISK: No validation
func GetUser(id string) (*User, error) {
    return db.Query("SELECT * FROM users WHERE id = " + id)  // SQL injection
}

// ✅ CORRECT: Validated and parameterized
func GetUser(id string) (*User, error) {
    if !isValidUUID(id) {
        return nil, ErrInvalidInput
    }
    return db.Query("SELECT * FROM users WHERE id = ?", id)
}
```

---

### 2. Authentication & Authorization

**Checklist**:
- [ ] Authentication required for sensitive operations
- [ ] Authorization checks before resource access
- [ ] Session tokens properly validated
- [ ] No hardcoded credentials
- [ ] Password storage uses bcrypt/scrypt (not MD5/SHA1)

**Common Vulnerabilities**:
```go
// ❌ CRITICAL: No authorization check
func DeleteUser(ctx context.Context, userID string) error {
    return db.Delete(userID)  // Anyone can delete any user
}

// ✅ CORRECT: Authorization verified
func DeleteUser(ctx context.Context, userID string) error {
    currentUser := auth.GetUser(ctx)
    if !currentUser.CanDelete(userID) {
        return ErrUnauthorized
    }
    return db.Delete(userID)
}
```

---

### 3. Cryptographic Usage

**Checklist**:
- [ ] Strong algorithms used (AES-256, RSA-2048+)
- [ ] No deprecated algorithms (MD5, SHA1 for security)
- [ ] Random number generation uses crypto/rand (not math/rand)
- [ ] TLS 1.2+ enforced for network communication
- [ ] Keys stored securely (environment variables, secret managers)

**Common Vulnerabilities**:
```go
// ❌ CRITICAL: Weak random for security token
import "math/rand"
func GenerateToken() string {
    return fmt.Sprintf("%d", rand.Int())  // Predictable
}

// ✅ CORRECT: Cryptographically secure random
import "crypto/rand"
func GenerateToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

---

### 4. Injection Prevention

**Checklist**:
- [ ] SQL queries use parameterization (no string concatenation)
- [ ] Command execution sanitized (avoid os/exec with user input)
- [ ] Template injection prevented (proper escaping)
- [ ] LDAP/XPath queries parameterized
- [ ] NoSQL injection prevented

**Common Vulnerabilities**:
```go
// ❌ HIGH RISK: Command injection
func RunScript(script string) error {
    cmd := exec.Command("sh", "-c", script)  // User-controlled
    return cmd.Run()
}

// ✅ CORRECT: Avoid shell execution with user input
func RunAllowedScript(scriptName string) error {
    allowed := map[string]bool{"backup": true, "cleanup": true}
    if !allowed[scriptName] {
        return ErrInvalidScript
    }
    cmd := exec.Command("./scripts/" + scriptName)
    return cmd.Run()
}
```

---

### 5. Sensitive Data Handling

**Checklist**:
- [ ] Secrets not hardcoded in source
- [ ] Passwords not logged or displayed
- [ ] PII encrypted at rest and in transit
- [ ] Sensitive data not in error messages
- [ ] Debug logs don't expose credentials

**Common Vulnerabilities**:
```go
// ❌ HIGH RISK: Logging sensitive data
func Login(username, password string) error {
    log.Printf("Login attempt: %s / %s", username, password)  // Logs password!
    return auth.Verify(username, password)
}

// ✅ CORRECT: No sensitive data in logs
func Login(username, password string) error {
    log.Printf("Login attempt: user=%s", username)  // No password
    return auth.Verify(username, password)
}
```

---

### 6. Network Security

**Checklist**:
- [ ] TLS properly configured (no InsecureSkipVerify)
- [ ] Certificate validation enabled
- [ ] Secure ciphers used (no weak ciphers)
- [ ] HTTP endpoints redirect to HTTPS
- [ ] CORS properly configured

**Common Vulnerabilities**:
```go
// ❌ CRITICAL: TLS verification disabled
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,  // Allows MITM attacks
        },
    },
}

// ✅ CORRECT: TLS properly verified
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,  // Enforce TLS 1.2+
        },
    },
}
```

---

### 7. Race Conditions & Concurrency

**Checklist**:
- [ ] Shared data protected with mutexes/channels
- [ ] No TOCTOU (Time-of-Check-Time-of-Use) vulnerabilities
- [ ] Atomic operations used for counters
- [ ] Context cancellation properly handled
- [ ] No goroutine leaks

**Common Vulnerabilities**:
```go
// ❌ MEDIUM RISK: Race condition
var balance int
func Withdraw(amount int) {
    if balance >= amount {  // Check
        balance -= amount   // Use (race window here)
    }
}

// ✅ CORRECT: Mutex protection
var balance int
var mu sync.Mutex
func Withdraw(amount int) {
    mu.Lock()
    defer mu.Unlock()
    if balance >= amount {
        balance -= amount
    }
}
```

---

## Review Template

Use this template during Layer 3 security assessment:

```markdown
## Security Assessment

**Risk Level**: [None/Low/Medium/High/Critical]

### Input Validation
- ✅ All external inputs validated (e.g., UUID format check at line 42)
- ⚠️ Missing length validation on `name` field (line 67)

### Authentication & Authorization
- ✅ Authentication required for all endpoints
- ✅ Authorization checks before resource access

### Cryptographic Usage
- ✅ Uses crypto/rand for token generation
- ✅ TLS 1.2+ enforced

### Injection Prevention
- ✅ SQL queries parameterized (GORM)
- ✅ No command execution with user input

### Sensitive Data Handling
- ✅ No credentials in source code
- ⚠️ Password visible in error logs (line 123)

### Network Security
- ✅ TLS properly configured
- ✅ Certificate validation enabled

### Concurrency Safety
- ✅ Shared cache protected with proper locking
- ✅ No obvious race conditions

**Identified Vulnerabilities**:
1. ⚠️ MEDIUM: Password in error logs (line 123)
   - Impact: Credentials exposed in log aggregation systems
   - Fix: Remove password from error message

2. ⚠️ LOW: Missing length validation on name field (line 67)
   - Impact: Potential buffer issues with extremely long inputs
   - Fix: Add max length check (e.g., 255 characters)

**Overall Assessment**: Low risk with 2 minor issues requiring fixes.
```

---

## Security Tools Integration

### Recommended Tools
1. **gosec** - Go security checker
   ```bash
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   gosec ./...
   ```

2. **staticcheck** - Static analysis
   ```bash
   go install honnef.co/go/tools/cmd/staticcheck@latest
   staticcheck ./...
   ```

3. **go-ruleguard** - Custom security rules
   ```bash
   go install github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest
   ```

---

## OWASP Top 10 for Go

### Quick Reference
1. **Injection**: Use parameterized queries
2. **Broken Authentication**: Implement proper session management
3. **Sensitive Data Exposure**: Encrypt PII, use HTTPS
4. **XML External Entities**: Disable external entity processing
5. **Broken Access Control**: Verify authorization for every request
6. **Security Misconfiguration**: Secure defaults, disable debug in prod
7. **XSS**: Escape output in templates (html/template auto-escapes)
8. **Insecure Deserialization**: Validate input before unmarshal
9. **Using Components with Known Vulnerabilities**: Run `go mod tidy`, audit deps
10. **Insufficient Logging**: Log security events, but not sensitive data

---

## References

- [Go Security Best Practices](https://github.com/golang/go/wiki/Security)
- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_Security_Cheat_Sheet.html)
- [gosec - Go Security Checker](https://github.com/securego/gosec)
