# Cryptography & Secrets Management

## When This Applies

- Encrypting sensitive data at rest or in transit
- Storing API keys, passwords, or credentials
- Generating random tokens or IDs
- Implementing TLS/HTTPS
- Signing or verifying data
- Key management and rotation
- Setting security HTTP headers

## Key Rules

### MUST Do (Cryptography)

- Use AES-256-GCM for symmetric encryption (authenticated encryption)
- Use Curve25519 or RSA-2048+ for asymmetric encryption
- Use cryptographically secure random number generators (CSPRNG)
- Generate unique IVs/nonces for each encryption operation
- Use authenticated encryption modes (GCM, CCM) when possible
- Use constant-time comparison for secrets
- Hash passwords with Argon2id, bcrypt, or scrypt (see authentication.md)

### MUST Do (Secrets Management)

- Store secrets in dedicated secrets management systems (use credstash)
- Never hardcode secrets in source code
- Never commit secrets to version control
- Use environment variables or secrets files (excluded from git)
- Rotate secrets regularly and after any suspected compromise
- Use short-lived, scoped credentials where possible
- Encrypt secrets at rest and in transit

### MUST NOT Do

- Use MD5, SHA1, or DES for security purposes
- Use ECB mode for block ciphers
- Reuse IVs/nonces with the same key
- Use `Math.random()` or similar for security-sensitive operations
- Store encryption keys alongside encrypted data
- Log secrets or include them in error messages
- Hardcode secrets in code, configs, or Dockerfiles

### SHOULD Do

- Use TLS 1.3 (or TLS 1.2 minimum)
- Set HSTS headers for HTTPS enforcement
- Implement Content Security Policy (CSP)
- Use secure cookie flags (Secure, HttpOnly, SameSite)
- Automate secret rotation
- Use credstash for key storage in production

## Safe Patterns

### Encryption

#### Python

```python
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import os
import secrets

# SECURITY: Generate a random encryption key
def generate_key() -> bytes:
    return secrets.token_bytes(32)  # 256 bits for AES-256

# SECURITY: AES-GCM encryption (recommended)
def encrypt_aes_gcm(plaintext: bytes, key: bytes) -> bytes:
    # SECURITY: Generate unique nonce for each encryption
    nonce = os.urandom(12)  # 96 bits for GCM
    aesgcm = AESGCM(key)
    ciphertext = aesgcm.encrypt(nonce, plaintext, associated_data=None)
    # Return nonce + ciphertext (nonce needed for decryption)
    return nonce + ciphertext

def decrypt_aes_gcm(encrypted: bytes, key: bytes) -> bytes:
    nonce = encrypted[:12]
    ciphertext = encrypted[12:]
    aesgcm = AESGCM(key)
    return aesgcm.decrypt(nonce, ciphertext, associated_data=None)

# SECURITY: Simple symmetric encryption with Fernet (AES-128-CBC + HMAC)
# NOTE: Fernet requires a URL-safe base64-encoded 32-byte key
def generate_fernet_key() -> bytes:
    return Fernet.generate_key()  # Returns base64-encoded key

def encrypt_fernet(data: bytes, key: bytes) -> bytes:
    # key must be from Fernet.generate_key(), not raw bytes
    f = Fernet(key)
    return f.encrypt(data)
```

#### Go

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

// SECURITY: AES-GCM encryption
func EncryptAESGCM(plaintext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    // SECURITY: Generate unique nonce for each encryption
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    // Nonce is prepended to ciphertext
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func DecryptAESGCM(ciphertext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

#### JavaScript/Node.js

```javascript
import crypto from 'crypto';

// SECURITY: AES-256-GCM encryption
function encryptAESGCM(plaintext, key) {
    // SECURITY: Generate unique IV for each encryption
    const iv = crypto.randomBytes(12);  // 96 bits for GCM
    
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    // SECURITY: Get auth tag for integrity verification
    const authTag = cipher.getAuthTag();
    
    // Return IV + authTag + ciphertext
    return Buffer.concat([iv, authTag, Buffer.from(encrypted, 'hex')]);
}

function decryptAESGCM(encrypted, key) {
    const iv = encrypted.slice(0, 12);
    const authTag = encrypted.slice(12, 28);
    const ciphertext = encrypted.slice(28);
    
    const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv);
    decipher.setAuthTag(authTag);
    
    let decrypted = decipher.update(ciphertext, null, 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
}

// SECURITY: Generate cryptographically secure random key
function generateKey() {
    return crypto.randomBytes(32);  // 256 bits
}
```

### Secure Random Generation

```python
import secrets
import os

# SECURITY: Use secrets module for security-sensitive randomness
token = secrets.token_hex(32)        # 64-char hex string
url_safe_token = secrets.token_urlsafe(32)  # URL-safe base64
random_bytes = secrets.token_bytes(32)  # Raw bytes
random_int = secrets.randbelow(1000)    # Random int [0, 1000)

# SECURITY: Alternative using os.urandom
random_bytes = os.urandom(32)

# DANGER: Never use these for security purposes
# import random
# random.random()  # VULNERABLE - predictable
# random.randint(0, 100)  # VULNERABLE
```

```javascript
import crypto from 'crypto';

// SECURITY: Use crypto module for security-sensitive randomness
const randomBytes = crypto.randomBytes(32);
const randomHex = crypto.randomBytes(32).toString('hex');
const randomUUID = crypto.randomUUID();

// DANGER: Never use for security purposes
// Math.random()  // VULNERABLE - predictable
```

```go
import "crypto/rand"

// SECURITY: Use crypto/rand for security-sensitive randomness
randomBytes := make([]byte, 32)
_, err := rand.Read(randomBytes)

// DANGER: Never use for security purposes
// import "math/rand"
// rand.Intn(100)  // VULNERABLE - predictable
```

### Secrets Management

#### Environment Variables (Development)

```python
import os
from dotenv import load_dotenv

# SECURITY: Load from .env file (not committed to git)
load_dotenv()

# SECURITY: Access secrets from environment
api_key = os.environ.get('API_KEY')
db_password = os.environ.get('DATABASE_PASSWORD')

# SECURITY: Fail fast if required secrets missing
if not api_key:
    raise RuntimeError("API_KEY environment variable required")
```

```bash
# .env file (add to .gitignore!)
API_KEY=sk-xxxxxxxxxxxx
DATABASE_PASSWORD=secure_password_here
```

```gitignore
# .gitignore - SECURITY: Never commit secrets
.env
.env.local
.env.*.local
*.pem
*.key
secrets/
```

#### Cloud Secrets Manager

```python
# AWS Secrets Manager
import boto3
import json

def get_secret(secret_name: str) -> dict:
    client = boto3.client('secretsmanager')
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response['SecretString'])

# Usage
secrets = get_secret('my-app/production')
db_password = secrets['database_password']
```

```javascript
// AWS Secrets Manager (Node.js)
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';

async function getSecret(secretName) {
    const client = new SecretsManagerClient();
    const response = await client.send(
        new GetSecretValueCommand({ SecretId: secretName })
    );
    return JSON.parse(response.SecretString);
}
```

#### Credstash V2 (Razorpay)

Credstash is the primary secrets management system for applications across development and production environments.

**Architecture:**
- **Credstash UI**: Utility tool to store secret values in DynamoDB tables
- **Kubestash**: Pulls secrets from DynamoDB `kubestash-*` tables and pushes them to Kubernetes secrets (runs every 10 minutes)

**Secrets Convention:**

Format: `namespace/secret-name/KEY`

- `KEY` must be in CAPITAL LETTERS
- Format must match regex: `^[a-zA-Z0-9\/_.-]*$`
- Keys in kubestash are case-insensitive (converted to uppercase before injection)
- Keys in credstash are case-sensitive

**Environment Endpoints:**

| Environment | URL |
|-------------|-----|
| Stage | `https://credstash-ui.concierge.stage.razorpay.in/dist/` |
| RSPL (Prod) | `https://credstash-ui.razorpay.com/dist/` |
| RSPL (Singapore) | `https://credstash-sg-ui.razorpay.com/dist/` |
| RZPX (Prod) | `https://credstash-ui.razorpayx.com/dist/` |
| DE Cluster (Prod) | `https://credstash-ui.de.razorpay.com/dist/` |
| Wallet Cluster (Prod) | `https://credstash-ui.razorpaywallet.com/dist/` |
| RTSPL Capital (Prod) | `https://credstash-capital-ui.razorpay.com/dist/` |

**Access Control:**
- Production: Access via Google Groups (Credstash Developers, Tech Leads, Managers, DevOps)
- Stage: Access via `developers@razorpay.com` and Google Groups
- Access requests handled via IT Team JIRA

**Usage Example:**

```bash
# Secret key format
namespace/secret-name/DATABASE_PASSWORD
namespace/secret-name/API_KEY

# Example for a service
payments/api-service/STRIPE_SECRET_KEY
```

> **Note:** When reading secrets, values are hashed to base64.

### Secure HTTP Headers

```python
# FastAPI/Starlette
from starlette.middleware import Middleware
from starlette.middleware.httpsredirect import HTTPSRedirectMiddleware

# SECURITY: Security headers middleware
@app.middleware("http")
async def add_security_headers(request, call_next):
    response = await call_next(request)
    
    # SECURITY: Prevent clickjacking
    response.headers["X-Frame-Options"] = "DENY"
    
    # SECURITY: Prevent MIME sniffing
    response.headers["X-Content-Type-Options"] = "nosniff"
    
    # SECURITY: Enable XSS filter
    response.headers["X-XSS-Protection"] = "1; mode=block"
    
    # SECURITY: HSTS - enforce HTTPS
    response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
    
    # SECURITY: Content Security Policy
    response.headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self'"
    
    return response
```

```javascript
// Express.js with helmet
import helmet from 'helmet';

// SECURITY: Apply security headers
app.use(helmet({
    contentSecurityPolicy: {
        directives: {
            defaultSrc: ["'self'"],
            scriptSrc: ["'self'"],
            styleSrc: ["'self'", "'unsafe-inline'"],
            imgSrc: ["'self'", "data:"],
        },
    },
    hsts: {
        maxAge: 31536000,
        includeSubDomains: true,
    },
}));
```

### Constant-Time Comparison

```python
import hmac

# SECURITY: Constant-time comparison prevents timing attacks
def secure_compare(a: str, b: str) -> bool:
    return hmac.compare_digest(a.encode(), b.encode())

# Usage for API key validation
if secure_compare(provided_key, expected_key):
    # Valid
    pass
```

```javascript
import crypto from 'crypto';

// SECURITY: Constant-time comparison
function secureCompare(a, b) {
    if (a.length !== b.length) {
        return false;
    }
    return crypto.timingSafeEqual(Buffer.from(a), Buffer.from(b));
}
```

## Unsafe Patterns (Flag These)

### Weak Cryptography
```python
# DANGER: MD5 is broken for security
import hashlib
hashlib.md5(data)  # VULNERABLE

# DANGER: DES is broken
from Crypto.Cipher import DES  # VULNERABLE

# DANGER: ECB mode leaks patterns
cipher = AES.new(key, AES.MODE_ECB)  # VULNERABLE
```

### Insecure Random
```python
# DANGER: Predictable random
import random
token = random.randint(0, 1000000)  # VULNERABLE
```

```javascript
// DANGER: Predictable random
const token = Math.random().toString(36);  // VULNERABLE
```

### Hardcoded Secrets
```python
# DANGER: Hardcoded secrets
API_KEY = "sk-1234567890"  # VULNERABLE
DB_PASSWORD = "password123"  # VULNERABLE
```

### Secrets in Logs
```python
# DANGER: Logging secrets
logger.info(f"Connecting with password: {password}")  # VULNERABLE
```

## Security Checklist

After implementing cryptography or secrets:

- [ ] Using AES-256-GCM or equivalent authenticated encryption
- [ ] Unique IVs/nonces for every encryption operation
- [ ] Cryptographically secure random number generation
- [ ] Secrets stored in secrets manager, not in code
- [ ] No secrets in version control
- [ ] Secrets not logged or included in errors
- [ ] Constant-time comparison for secret validation
- [ ] HTTPS enforced with TLS 1.2+
- [ ] Security headers configured (HSTS, CSP, X-Frame-Options)
- [ ] Secure cookie flags set

## Source References

- [Cryptographic Storage Cheat Sheet](../../security-guidelines/Cryptographic_Storage_Cheat_Sheet.md)
- [Secrets Management Cheat Sheet](../../security-guidelines/Secrets_Management_Cheat_Sheet.md)
- [Key Management Cheat Sheet](../../security-guidelines/Key_Management_Cheat_Sheet.md)
- [Transport Layer Security Cheat Sheet](../../security-guidelines/Transport_Layer_Security_Cheat_Sheet.md)
- [HTTP Headers Cheat Sheet](../../security-guidelines/HTTP_Headers_Cheat_Sheet.md)
- [HTTP Strict Transport Security Cheat Sheet](../../security-guidelines/HTTP_Strict_Transport_Security_Cheat_Sheet.md)
- [Content Security Policy Cheat Sheet](../../security-guidelines/Content_Security_Policy_Cheat_Sheet.md)

