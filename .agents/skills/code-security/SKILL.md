---
name: application-security
description: Apply security best practices when writing, reviewing, or discussing code. Covers authentication, injection prevention, API security, input validation, infrastructure, and AI/LLM security for Python, Go, JS/TS, React, PHP, Node.js.
---

# Application Security

This skill provides comprehensive security guidance for application development, synthesized from OWASP security cheat sheets covering 78 security topics.

## Triggers

Activate this skill when:

- Writing code that handles user input, authentication, database queries, file operations, or external APIs
- Reviewing existing code for security vulnerabilities
- Discussing architecture or design patterns
- Vibe coding sessions (embed security by default - speed should not compromise safety)
- Choosing dependencies or third-party libraries
- Implementing any feature that touches sensitive data

## Behavior Rules

1. **Proactively flag security issues** - Before writing risky code, warn about potential vulnerabilities
2. **Add inline security comments** - Use `// SECURITY:` or `# SECURITY:` comments to explain why specific patterns are used
3. **Provide security summary** - After generating significant code blocks, include a quick checklist of security considerations addressed
4. **Suggest specific fixes** - When reviewing code, provide before/after examples with explanations

## Quick Reference - Critical Security Rules

### MUST Always Do

- Use parameterized queries for ALL database operations (never concatenate user input into SQL)
- Validate and sanitize ALL user input on the server side
- Use HTTPS/TLS for all network communications
- Hash passwords with bcrypt, Argon2, or scrypt (never MD5/SHA1)
- Implement proper authentication and session management
- Apply the principle of least privilege
- Encode output based on context (HTML, JavaScript, URL, CSS)
- Use CSRF tokens for state-changing operations
- Set secure HTTP headers (CSP, HSTS, X-Frame-Options, etc.)
- Log security-relevant events (but never log sensitive data like passwords)

### MUST NEVER Do

- Trust user input without validation
- Store passwords in plaintext or with weak hashing
- Expose sensitive data in error messages, logs, or responses
- Use `eval()` or dynamic code execution with user input
- Disable SSL/TLS certificate verification
- Hardcode secrets, API keys, or credentials in source code
- Use outdated or vulnerable dependencies
- Expose internal system details in production errors
- Allow unrestricted file uploads
- Use `dangerouslySetInnerHTML` (React) or equivalent without sanitization

### SHOULD Follow

- Implement rate limiting on authentication endpoints
- Use Content Security Policy (CSP) headers
- Implement proper session timeout and rotation
- Use security linters and static analysis tools
- Keep dependencies updated and scan for vulnerabilities
- Implement proper logging and monitoring
- Use prepared statements even when you "know" the input is safe
- Apply defense in depth - multiple layers of security

## Security Comment Format

When generating code, add inline comments like:

```python
# SECURITY: Using parameterized query to prevent SQL injection
cursor.execute("SELECT * FROM users WHERE id = %s", (user_id,))
```

```javascript
// SECURITY: Escaping HTML to prevent XSS
const safeContent = DOMPurify.sanitize(userInput);
```

```go
// SECURITY: Using filepath.Clean to prevent path traversal
cleanPath := filepath.Clean(userInput)
```

## Security Summary Template

After generating significant code, provide:

```
## Security Checklist
- [ ] Input validation implemented
- [ ] Output encoding applied
- [ ] Authentication/authorization checked
- [ ] Sensitive data protected
- [ ] Error handling doesn't leak info
- [ ] Logging implemented (no sensitive data)
```

## Resources

For detailed patterns and examples, see:

- `resources/authentication.md` - Auth flows, passwords, MFA, sessions, authorization, OAuth2
- `resources/injection-prevention.md` - SQL injection, XSS, CSRF, command injection, clickjacking
- `resources/api-security.md` - REST, GraphQL, gRPC, WebSocket, microservices security
- `resources/input-validation.md` - File uploads, deserialization, SSRF, XXE prevention
- `resources/infrastructure.md` - Docker, Kubernetes, cloud, CI/CD, supply chain security
- `resources/ai-llm-security.md` - Prompt injection, AI agent security, model ops
- `resources/cryptography-secrets.md` - Encryption, TLS, HTTP headers, secrets management
- `resources/secure-development.md` - Code review, threat modeling, logging, error handling
- `resources/framework-specific.md` - Node.js, Django, Laravel, Rails, PHP security

