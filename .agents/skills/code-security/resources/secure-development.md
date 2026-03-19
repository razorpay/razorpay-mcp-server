# Secure Development Practices

## When This Applies

- Designing new features or systems
- Writing error handling and logging
- Conducting code reviews
- Setting up development workflows
- Planning security testing
- Implementing observability

## Key Rules

### MUST Do (Error Handling)

- Catch and handle all exceptions appropriately
- Return generic error messages to users
- Log detailed errors server-side (without sensitive data)
- Use appropriate HTTP status codes
- Implement graceful degradation
- Have fallback behavior for external service failures

### MUST Do (Logging)

- Log security-relevant events (authentication, authorization, access)
- Include timestamp, user ID, action, and outcome
- Use structured logging (JSON) for easy analysis
- Protect log files from unauthorized access
- Implement log retention policies

### MUST NOT Do (Error Handling)

- Expose stack traces to end users
- Include sensitive data in error messages
- Reveal internal system details (paths, versions, configs)
- Log passwords, tokens, or PII
- Ignore or swallow exceptions silently

### MUST NOT Do (Logging)

- Log passwords, API keys, or tokens
- Log full credit card numbers or SSNs
- Log session IDs in a way that enables hijacking
- Store logs indefinitely without rotation

### SHOULD Do

- Perform threat modeling early in design
- Conduct security code reviews
- Use security linters and SAST tools
- Implement rate limiting for sensitive operations
- Use feature flags for gradual rollouts
- Have incident response procedures documented

## Safe Patterns

### Error Handling

#### Python

```python
import logging
from typing import Optional

logger = logging.getLogger(__name__)

class AppError(Exception):
    """Base application error with safe user message."""
    def __init__(self, message: str, user_message: str = "An error occurred"):
        super().__init__(message)
        self.user_message = user_message  # Safe to show users


def process_request(user_id: str, data: dict) -> dict:
    try:
        result = perform_operation(data)
        return {"success": True, "data": result}
    
    except ValidationError as e:
        # SECURITY: Log details internally, return safe message
        logger.warning(f"Validation error for user {user_id}: {e}")
        return {"success": False, "error": "Invalid input provided"}
    
    except PermissionError as e:
        # SECURITY: Log the attempt
        logger.warning(f"Unauthorized access attempt by user {user_id}: {e}")
        return {"success": False, "error": "Access denied"}
    
    except Exception as e:
        # SECURITY: Log full details, return generic message
        logger.exception(f"Unexpected error for user {user_id}")
        return {"success": False, "error": "An unexpected error occurred"}


# FastAPI example
from fastapi import HTTPException, Request
from fastapi.responses import JSONResponse

@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    # SECURITY: Log details, return generic error
    logger.exception(f"Unhandled exception: {request.url}")
    return JSONResponse(
        status_code=500,
        content={"error": "An internal error occurred"}
    )
```

#### Go

```go
import (
    "log"
    "net/http"
)

// SECURITY: Custom error type with safe user message
type AppError struct {
    Code       int
    Message    string  // Internal message for logging
    UserMsg    string  // Safe message for users
}

func (e *AppError) Error() string {
    return e.Message
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    result, err := processData(r)
    if err != nil {
        handleError(w, err)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}

func handleError(w http.ResponseWriter, err error) {
    // SECURITY: Log full error details server-side
    log.Printf("Error: %v", err)
    
    // SECURITY: Return safe message to user
    if appErr, ok := err.(*AppError); ok {
        http.Error(w, appErr.UserMsg, appErr.Code)
        return
    }
    
    // Generic fallback for unknown errors
    http.Error(w, "An error occurred", http.StatusInternalServerError)
}
```

#### JavaScript/Node.js

```javascript
import winston from 'winston';

const logger = winston.createLogger({
    level: 'info',
    format: winston.format.json(),
    transports: [
        new winston.transports.File({ filename: 'error.log', level: 'error' }),
        new winston.transports.File({ filename: 'combined.log' }),
    ],
});

// SECURITY: Error handling middleware for Express
function errorHandler(err, req, res, next) {
    // SECURITY: Log full error details
    logger.error({
        message: err.message,
        stack: err.stack,
        url: req.url,
        method: req.method,
        userId: req.user?.id,
        timestamp: new Date().toISOString()
    });
    
    // SECURITY: Return safe message to client
    const statusCode = err.statusCode || 500;
    const userMessage = err.isOperational 
        ? err.message 
        : 'An unexpected error occurred';
    
    res.status(statusCode).json({
        error: userMessage
        // SECURITY: Never include stack trace in production
    });
}

// Custom error class
class AppError extends Error {
    constructor(message, statusCode = 500, isOperational = true) {
        super(message);
        this.statusCode = statusCode;
        this.isOperational = isOperational;  // Safe to show to users
    }
}
```

### Security Logging

```python
import logging
import json
from datetime import datetime
from typing import Optional

# SECURITY: Structured security logger
class SecurityLogger:
    def __init__(self):
        self.logger = logging.getLogger('security')
        handler = logging.FileHandler('security.log')
        handler.setFormatter(logging.Formatter('%(message)s'))
        self.logger.addHandler(handler)
        self.logger.setLevel(logging.INFO)
    
    def log_event(
        self,
        event_type: str,
        user_id: Optional[str],
        action: str,
        outcome: str,
        details: dict = None,
        ip_address: str = None
    ):
        event = {
            'timestamp': datetime.utcnow().isoformat(),
            'event_type': event_type,
            'user_id': user_id,
            'action': action,
            'outcome': outcome,
            'ip_address': ip_address,
            'details': details or {}
        }
        self.logger.info(json.dumps(event))
    
    def log_auth_success(self, user_id: str, ip: str, method: str = 'password'):
        self.log_event('authentication', user_id, 'login', 'success',
                      {'method': method}, ip)
    
    def log_auth_failure(self, username: str, ip: str, reason: str):
        # SECURITY: Log attempts but don't confirm if user exists
        self.log_event('authentication', None, 'login', 'failure',
                      {'reason': reason, 'attempted_user': username}, ip)
    
    def log_access_denied(self, user_id: str, resource: str, ip: str):
        self.log_event('authorization', user_id, f'access:{resource}', 'denied',
                      {}, ip)


# Usage
security_log = SecurityLogger()

# On successful login
security_log.log_auth_success(user.id, request.client.host)

# On failed login
security_log.log_auth_failure(username, request.client.host, 'invalid_password')

# On authorization failure
security_log.log_access_denied(user.id, '/admin/users', request.client.host)
```

### Data Sanitization for Logs

```python
import re
from typing import Any

class LogSanitizer:
    """Sanitize sensitive data before logging."""
    
    SENSITIVE_PATTERNS = {
        'password': r'password["\s:=]+["\']?[\w@#$%^&*]+',
        'token': r'(bearer\s+)?[a-zA-Z0-9_-]{20,}',
        'credit_card': r'\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b',
        'ssn': r'\b\d{3}-\d{2}-\d{4}\b',
        'api_key': r'(api[_-]?key|apikey)["\s:=]+["\']?[\w-]+',
    }
    
    SENSITIVE_FIELDS = {'password', 'token', 'api_key', 'secret', 'credit_card'}
    
    @classmethod
    def sanitize_string(cls, text: str) -> str:
        """Redact sensitive patterns from string."""
        result = text
        for name, pattern in cls.SENSITIVE_PATTERNS.items():
            result = re.sub(pattern, f'[REDACTED:{name}]', result, flags=re.IGNORECASE)
        return result
    
    @classmethod
    def sanitize_dict(cls, data: dict) -> dict:
        """Recursively sanitize sensitive fields in dict."""
        result = {}
        for key, value in data.items():
            if key.lower() in cls.SENSITIVE_FIELDS:
                result[key] = '[REDACTED]'
            elif isinstance(value, dict):
                result[key] = cls.sanitize_dict(value)
            elif isinstance(value, str):
                result[key] = cls.sanitize_string(value)
            else:
                result[key] = value
        return result


# Usage
sanitizer = LogSanitizer()
safe_data = sanitizer.sanitize_dict(request_data)
logger.info(f"Request: {safe_data}")
```

### Threat Modeling Considerations

When designing features, consider:

```markdown
## STRIDE Threat Model

| Threat | Description | Example Mitigations |
|--------|-------------|---------------------|
| **Spoofing** | Impersonating users/systems | Strong authentication, MFA |
| **Tampering** | Modifying data/code | Integrity checks, signed commits |
| **Repudiation** | Denying actions | Audit logging, non-repudiation |
| **Information Disclosure** | Data leaks | Encryption, access controls |
| **Denial of Service** | Making service unavailable | Rate limiting, resource limits |
| **Elevation of Privilege** | Gaining unauthorized access | Least privilege, input validation |
```

### Code Review Security Checklist

```markdown
## Security Code Review Checklist

### Authentication & Authorization
- [ ] Authentication required for protected endpoints
- [ ] Authorization checks on every request
- [ ] Object-level permissions validated

### Input Handling
- [ ] All input validated and sanitized
- [ ] Parameterized queries used
- [ ] File uploads validated by content

### Output
- [ ] Output encoded for context
- [ ] Error messages don't leak sensitive info
- [ ] No sensitive data in responses

### Cryptography
- [ ] Strong algorithms used (AES-256, bcrypt)
- [ ] Secrets not hardcoded
- [ ] Secure random generation

### Logging & Error Handling
- [ ] Security events logged
- [ ] No sensitive data in logs
- [ ] Exceptions handled gracefully
```

### Rate Limiting

```python
from functools import wraps
from collections import defaultdict
import time

class RateLimiter:
    """Simple in-memory rate limiter."""
    
    def __init__(self, max_requests: int, window_seconds: int):
        self.max_requests = max_requests
        self.window = window_seconds
        self.requests = defaultdict(list)
    
    def is_allowed(self, key: str) -> bool:
        now = time.time()
        window_start = now - self.window
        
        # Clean old entries
        self.requests[key] = [t for t in self.requests[key] if t > window_start]
        
        if len(self.requests[key]) >= self.max_requests:
            return False
        
        self.requests[key].append(now)
        return True


# SECURITY: Rate limit decorator
login_limiter = RateLimiter(max_requests=5, window_seconds=60)

def rate_limit(limiter: RateLimiter, key_func):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            key = key_func(*args, **kwargs)
            if not limiter.is_allowed(key):
                raise HTTPException(429, "Too many requests")
            return await func(*args, **kwargs)
        return wrapper
    return decorator

@rate_limit(login_limiter, lambda request, **_: request.client.host)
async def login(request: Request, credentials: LoginRequest):
    # ... login logic
```

## Unsafe Patterns (Flag These)

### Exposing Stack Traces
```python
# DANGER: Exposing internals to users
@app.exception_handler(Exception)
async def handler(request, exc):
    return JSONResponse({
        "error": str(exc),
        "stack": traceback.format_exc()  # VULNERABLE
    })
```

### Logging Sensitive Data
```python
# DANGER: Logging passwords/tokens
logger.info(f"User login: {username}, password: {password}")  # VULNERABLE
logger.info(f"API call with key: {api_key}")  # VULNERABLE
```

### Swallowing Exceptions
```python
# DANGER: Silent failure hides issues
try:
    process_payment()
except Exception:
    pass  # VULNERABLE - silent failure
```

### Detailed Error Messages
```python
# DANGER: Leaking internal details
return {
    "error": f"Database connection failed: {db_host}:{db_port}",
    "query": sql_query,
    "version": "PostgreSQL 14.2"
}  # VULNERABLE
```

## Security Checklist

For secure development practices:

- [ ] Error messages don't expose internals
- [ ] Stack traces not shown to users
- [ ] Security events logged with context
- [ ] No sensitive data in logs
- [ ] Rate limiting on sensitive endpoints
- [ ] Code reviewed for security issues
- [ ] Threat model considered for new features
- [ ] Dependencies scanned for vulnerabilities
- [ ] Security testing included in CI/CD

## Source References

- [Secure Product Design Cheat Sheet](../../security-guidelines/Secure_Product_Design_Cheat_Sheet.md)
- [Secure Code Review Cheat Sheet](../../security-guidelines/Secure_Code_Review_Cheat_Sheet.md)
- [Threat Modeling Cheat Sheet](../../security-guidelines/Threat_Modeling_Cheat_Sheet.md)
- [Error Handling Cheat Sheet](../../security-guidelines/Error_Handling_Cheat_Sheet.md)
- [Logging Cheat Sheet](../../security-guidelines/Logging_Cheat_Sheet.md)
- [Attack Surface Analysis Cheat Sheet](../../security-guidelines/Attack_Surface_Analysis_Cheat_Sheet.md)
- [Abuse Case Cheat Sheet](../../security-guidelines/Abuse_Case_Cheat_Sheet.md)
- [Database Security Cheat Sheet](../../security-guidelines/Database_Security_Cheat_Sheet.md)

