# Application Security: Secure Coding Skill

A comprehensive Claude Skill for secure coding practices, synthesized from 78 OWASP security cheat sheets. This skill helps developers write secure code by providing real-time security guidance during coding, code review, and architecture discussions.

## Features

- **Proactive Security Guidance**: Claude flags security issues before writing risky code
- **Inline Security Comments**: Adds `// SECURITY:` comments explaining secure patterns
- **Security Checklists**: Provides post-generation checklists for verification
- **Multi-Language Support**: Examples in Python, Go, JavaScript/TypeScript, PHP, and Ruby
- **Framework Coverage**: Node.js, Django, Laravel, Rails, and PHP-specific guidance

## Triggers

The skill activates when:

- Writing code that handles user input, authentication, database queries, file operations, or external APIs
- Reviewing existing code for security vulnerabilities
- Discussing architecture or design patterns
- Vibe coding sessions (embeds security by default)
- Choosing dependencies or third-party libraries

## Structure

```
application-security/
├── Skill.md                    # Core skill - triggers, behavior rules, quick reference
└── resources/
    ├── authentication.md       # Auth, MFA, sessions, passwords, authorization
    ├── injection-prevention.md # SQL, XSS, CSRF, command injection
    ├── api-security.md         # REST, GraphQL, gRPC, WebSocket, JWT
    ├── input-validation.md     # File uploads, SSRF, XXE, deserialization
    ├── infrastructure.md       # Docker, K8s, CI/CD, cloud, supply chain
    ├── ai-llm-security.md      # Prompt injection, AI agent security
    ├── cryptography-secrets.md # Encryption, TLS, secrets management
    ├── secure-development.md   # Error handling, logging, code review
    └── framework-specific.md   # Node.js, Django, Laravel, Rails, PHP
```

## Topics Covered

| Resource | Key Topics |
|----------|------------|
| **authentication.md** | Password hashing (Argon2, bcrypt), session management, JWT validation, MFA, OAuth2, RBAC/ABAC |
| **injection-prevention.md** | SQL injection, XSS (DOM, stored, reflected), CSRF tokens, command injection, NoSQL injection |
| **api-security.md** | JWT security, rate limiting, CORS configuration, GraphQL depth limiting, WebSocket auth |
| **input-validation.md** | Allowlist validation, file upload security, SSRF prevention, XXE prevention, deserialization |
| **infrastructure.md** | Secure Dockerfiles, Kubernetes security contexts, CI/CD pipeline security, supply chain |
| **ai-llm-security.md** | Prompt injection defense, AI agent tool security, output validation, RAG poisoning |
| **cryptography-secrets.md** | AES-GCM encryption, secure random generation, secrets management, TLS, HTTP headers |
| **secure-development.md** | Error handling, security logging, threat modeling, code review checklists |
| **framework-specific.md** | Express/helmet, Django settings, Laravel validation, Rails strong params, PHP config |

## Installation

### For Claude.ai (Pro/Max/Team/Enterprise)

1. Create a ZIP file of the skill:
   ```bash
   cd skills
   zip -r application-security.zip application-security/
   ```

2. Go to **Settings > Capabilities > Skills**

3. Upload `application-security.zip`

4. Enable the skill

### For Claude Code (CLI)

Claude Code can use this skill by reading the files directly from your project or a central location.

#### Option 1: Add to Your Project

Copy the skill to your project's `.claude/skills/` directory:

```bash
# From your project root
mkdir -p .claude/skills
cp -r /path/to/application-security .claude/skills/
```

#### Option 2: Central Skills Location

Store skills in a central location and reference them:

```bash
# Create central skills directory
mkdir -p ~/.claude/skills
cp -r /path/to/application-security ~/.claude/skills/
```

#### Option 3: Use with CLAUDE.md

Reference the skill in your project's `CLAUDE.md` file:

```markdown
# Project Instructions

## Security
When writing code, follow the security guidelines in:
- `skills/application-security/Skill.md` for core rules
- `skills/application-security/resources/` for detailed patterns

Always add `// SECURITY:` comments explaining security decisions.
```

#### Using the Skill in Claude Code

Once set up, invoke security guidance:

```bash
# Ask Claude Code to review security
claude "Review this file for security issues using the application-security skill"

# Generate secure code
claude "Create a login endpoint following the authentication security guidelines"

# Check specific concerns
claude "Is this database query safe from SQL injection?"
```


### For Cursor IDE

Add the skill to your project and reference it in your `.cursorrules` file:

```bash
# Copy skill to your project
cp -r /path/to/application-security .cursor/skills/
```

Create or update `.cursorrules` in your project root:

```markdown
# Security Guidelines

When writing or reviewing code, follow the security best practices defined in:
- `.cursor/skills/application-security/Skill.md`

Key behaviors:
- Add `// SECURITY:` inline comments explaining security patterns
- Use parameterized queries for all database operations
- Validate all user input on the server side
- Never hardcode secrets - use environment variables
- Provide a security checklist after generating significant code

For detailed patterns, reference the resource files in `.cursor/skills/application-security/resources/`
```

## Usage Examples

Once enabled, Claude will automatically apply security best practices:

### When Writing Database Queries

```python
# Claude will generate:
# SECURITY: Using parameterized query to prevent SQL injection
cursor.execute("SELECT * FROM users WHERE email = %s", (email,))
```

### When Implementing Authentication

```javascript
// Claude will generate:
// SECURITY: Hash password with bcrypt, cost factor 12
const hash = await bcrypt.hash(password, 12);

// SECURITY: Constant-time comparison prevents timing attacks
const valid = await bcrypt.compare(inputPassword, hash);
```

### After Generating Code

Claude provides a security checklist:

```
## Security Checklist
- [x] Input validation implemented
- [x] Parameterized queries used
- [ ] Rate limiting (consider adding)
- [x] Error handling doesn't leak info
```

## Source Material

This skill synthesizes guidance from 78 OWASP Cheat Sheets including:

- Authentication, Session Management, Password Storage
- SQL Injection Prevention, XSS Prevention, CSRF Prevention
- REST Security, GraphQL Security, WebSocket Security
- Docker Security, Kubernetes Security, CI/CD Security
- LLM Prompt Injection Prevention, AI Agent Security
- Cryptographic Storage, Secrets Management, TLS
- And many more...

## Contributing

To update or extend this skill:

1. Add new patterns to the relevant resource file
2. Follow the existing format (Key Rules, Safe Patterns, Unsafe Patterns, Checklist)
3. Include code examples in multiple languages where applicable
4. Reference source OWASP cheat sheets


