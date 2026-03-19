# Input Validation & Data Handling Security

## When This Applies

- Processing any user-supplied input
- Handling file uploads
- Making HTTP requests to user-provided URLs
- Parsing XML, JSON, or other structured data
- Processing redirects or forwards
- Deserializing objects from untrusted sources

## Key Rules

### MUST Do

- Validate all input on the server side (never trust client-side validation alone)
- Use allowlist validation (define what IS allowed, reject everything else)
- Validate both syntactic (format) and semantic (business logic) correctness
- Limit input length, file size, and array sizes
- Validate file types by content (magic bytes), not just extension or Content-Type
- Store uploaded files outside the webroot
- Use allowlists for URLs when making server-side requests
- Disable external entity processing in XML parsers
- Disable following redirects for server-side HTTP requests

### MUST NOT Do

- Rely on denylist validation as primary defense
- Trust Content-Type headers for file validation
- Use user-supplied filenames directly
- Make HTTP requests to arbitrary user-supplied URLs
- Deserialize untrusted data with unsafe deserializers
- Parse XML with external entity processing enabled
- Allow path traversal characters in file paths (`../`)

### SHOULD Do

- Use JSON Schema or similar for structured input validation
- Generate new random filenames for uploaded files
- Scan uploaded files with antivirus when possible
- Use Content-Disarm-Reconstruct (CDR) for documents
- Validate redirects against an allowlist of domains
- Use type-safe parsing (parseInt, etc.) with error handling

## Safe Patterns

### Input Validation

#### Python

```python
from pydantic import BaseModel, EmailStr, Field, validator
import re

# SECURITY: Use Pydantic for strict input validation
class UserRegistration(BaseModel):
    username: str = Field(..., min_length=3, max_length=30, regex=r'^[a-zA-Z0-9_]+$')
    email: EmailStr
    age: int = Field(..., ge=13, le=120)
    
    @validator('username')
    def username_alphanumeric(cls, v):
        if not v.isalnum() and '_' not in v:
            raise ValueError('Username must be alphanumeric')
        return v

# SECURITY: Validate against fixed set of options
ALLOWED_ROLES = {'user', 'admin', 'moderator'}

def validate_role(role: str) -> str:
    if role not in ALLOWED_ROLES:
        raise ValueError(f"Invalid role. Must be one of: {ALLOWED_ROLES}")
    return role

# SECURITY: Validate numeric input with bounds
def validate_quantity(value: str) -> int:
    try:
        qty = int(value)
        if not (1 <= qty <= 100):
            raise ValueError("Quantity must be between 1 and 100")
        return qty
    except (ValueError, TypeError):
        raise ValueError("Invalid quantity")
```

#### Go

```go
import (
    "regexp"
    "errors"
)

// SECURITY: Validate input with strict patterns
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateUsername(username string) error {
    if !usernameRegex.MatchString(username) {
        return errors.New("invalid username format")
    }
    return nil
}

// SECURITY: Validate against allowlist
var allowedRoles = map[string]bool{
    "user": true, "admin": true, "moderator": true,
}

func ValidateRole(role string) error {
    if !allowedRoles[role] {
        return errors.New("invalid role")
    }
    return nil
}
```

#### JavaScript/TypeScript

```typescript
import { z } from 'zod';

// SECURITY: Use Zod for runtime validation
const UserSchema = z.object({
    username: z.string()
        .min(3)
        .max(30)
        .regex(/^[a-zA-Z0-9_]+$/, 'Username must be alphanumeric'),
    email: z.string().email(),
    age: z.number().int().min(13).max(120),
});

function validateUser(input: unknown) {
    // SECURITY: Parse throws on invalid input
    return UserSchema.parse(input);
}

// SECURITY: Validate against fixed options
const ALLOWED_ROLES = ['user', 'admin', 'moderator'] as const;
const RoleSchema = z.enum(ALLOWED_ROLES);

// SECURITY: Strict ID validation
const IdSchema = z.string().uuid();
```

### File Upload Security

#### Python (FastAPI)

```python
import os
import uuid
import magic
from pathlib import Path

# SECURITY: Allowlist of permitted extensions and MIME types
ALLOWED_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.pdf'}
ALLOWED_MIMETYPES = {
    'image/jpeg', 'image/png', 'image/gif', 'application/pdf'
}
MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB

async def secure_file_upload(file: UploadFile) -> str:
    # SECURITY: Validate file size
    contents = await file.read()
    if len(contents) > MAX_FILE_SIZE:
        raise HTTPException(400, "File too large")
    
    # SECURITY: Validate extension from original filename
    original_ext = Path(file.filename).suffix.lower()
    if original_ext not in ALLOWED_EXTENSIONS:
        raise HTTPException(400, "File type not allowed")
    
    # SECURITY: Validate actual file content (magic bytes)
    mime = magic.from_buffer(contents, mime=True)
    if mime not in ALLOWED_MIMETYPES:
        raise HTTPException(400, "Invalid file content")
    
    # SECURITY: Generate new random filename
    new_filename = f"{uuid.uuid4()}{original_ext}"
    
    # SECURITY: Store outside webroot
    upload_path = Path("/var/uploads") / new_filename
    
    # SECURITY: Prevent path traversal
    if not str(upload_path).startswith("/var/uploads/"):
        raise HTTPException(400, "Invalid path")
    
    with open(upload_path, 'wb') as f:
        f.write(contents)
    
    return new_filename
```

#### JavaScript/Node.js

```javascript
import multer from 'multer';
import path from 'path';
import crypto from 'crypto';
import { fileTypeFromBuffer } from 'file-type';

// SECURITY: File filter with strict validation
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'application/pdf'];
const MAX_SIZE = 10 * 1024 * 1024; // 10MB

const storage = multer.diskStorage({
    destination: '/var/uploads/',  // Outside webroot
    filename: (req, file, cb) => {
        // SECURITY: Generate random filename
        const ext = path.extname(file.originalname).toLowerCase();
        const randomName = crypto.randomBytes(16).toString('hex');
        cb(null, `${randomName}${ext}`);
    }
});

const upload = multer({
    storage,
    limits: { fileSize: MAX_SIZE },
    fileFilter: async (req, file, cb) => {
        // SECURITY: Check extension
        const ext = path.extname(file.originalname).toLowerCase();
        const allowedExts = ['.jpg', '.jpeg', '.png', '.gif', '.pdf'];
        
        if (!allowedExts.includes(ext)) {
            return cb(new Error('File type not allowed'));
        }
        
        cb(null, true);
    }
});

// SECURITY: Additional content validation after upload
async function validateFileContent(filePath) {
    const buffer = await fs.promises.readFile(filePath);
    const type = await fileTypeFromBuffer(buffer);
    
    if (!type || !ALLOWED_TYPES.includes(type.mime)) {
        await fs.promises.unlink(filePath);
        throw new Error('Invalid file content');
    }
}
```

### SSRF Prevention

#### Python

```python
from urllib.parse import urlparse
import ipaddress
import socket

# SECURITY: Allowlist of permitted domains
ALLOWED_HOSTS = {'api.trusted.com', 'data.trusted.com'}

def validate_url(url: str) -> str:
    parsed = urlparse(url)
    
    # SECURITY: Only allow HTTPS
    if parsed.scheme != 'https':
        raise ValueError("Only HTTPS URLs allowed")
    
    # SECURITY: Check against allowlist
    if parsed.hostname not in ALLOWED_HOSTS:
        raise ValueError("Host not in allowlist")
    
    return url

# SECURITY: For dynamic URLs, block internal networks
def is_internal_ip(hostname: str) -> bool:
    try:
        ip = socket.gethostbyname(hostname)
        ip_obj = ipaddress.ip_address(ip)
        
        # Block private, loopback, link-local addresses
        return (
            ip_obj.is_private or
            ip_obj.is_loopback or
            ip_obj.is_link_local or
            ip_obj.is_reserved
        )
    except socket.gaierror:
        return True  # Block if can't resolve

def safe_fetch(url: str):
    parsed = urlparse(url)
    
    if parsed.scheme not in ('http', 'https'):
        raise ValueError("Invalid scheme")
    
    # SECURITY: Block internal IPs
    if is_internal_ip(parsed.hostname):
        raise ValueError("Internal addresses not allowed")
    
    # SECURITY: Disable redirects to prevent bypass
    response = requests.get(url, allow_redirects=False, timeout=10)
    return response
```

#### JavaScript/Node.js

```javascript
import { URL } from 'url';
import dns from 'dns/promises';
import ipaddr from 'ipaddr.js';

// SECURITY: Allowlist of permitted hosts
const ALLOWED_HOSTS = new Set(['api.trusted.com', 'data.trusted.com']);

async function validateUrl(urlString) {
    const url = new URL(urlString);
    
    // SECURITY: Only allow HTTPS
    if (url.protocol !== 'https:') {
        throw new Error('Only HTTPS URLs allowed');
    }
    
    // SECURITY: Check allowlist
    if (!ALLOWED_HOSTS.has(url.hostname)) {
        throw new Error('Host not in allowlist');
    }
    
    return url;
}

// SECURITY: Check if IP is internal
async function isInternalAddress(hostname) {
    try {
        const addresses = await dns.resolve4(hostname);
        for (const addr of addresses) {
            const parsed = ipaddr.parse(addr);
            const range = parsed.range();
            
            if (['private', 'loopback', 'linkLocal', 'reserved'].includes(range)) {
                return true;
            }
        }
        return false;
    } catch {
        return true; // Block if can't resolve
    }
}
```

### XML Parsing (XXE Prevention)

#### Python

```python
from lxml import etree
import defusedxml.ElementTree as ET

# SECURITY: Use defusedxml for safe XML parsing
def safe_parse_xml(xml_string: str):
    # defusedxml blocks XXE by default
    return ET.fromstring(xml_string)

# SECURITY: If using lxml, disable external entities
def safe_lxml_parse(xml_string: str):
    parser = etree.XMLParser(
        resolve_entities=False,
        no_network=True,
        dtd_validation=False,
        load_dtd=False
    )
    return etree.fromstring(xml_string.encode(), parser)
```

#### JavaScript/Node.js

```javascript
import { XMLParser } from 'fast-xml-parser';

// SECURITY: Configure parser to prevent XXE
const parser = new XMLParser({
    // Disable external entity processing
    processEntities: false,
    // Additional safety options
    htmlEntities: false,
});

function safeParseXml(xmlString) {
    return parser.parse(xmlString);
}
```

### Path Traversal Prevention

```python
import os
from pathlib import Path

UPLOAD_DIR = Path("/var/uploads").resolve()

def safe_file_path(user_filename: str) -> Path:
    # SECURITY: Remove any path components
    safe_name = os.path.basename(user_filename)
    
    # SECURITY: Build path and resolve
    full_path = (UPLOAD_DIR / safe_name).resolve()
    
    # SECURITY: Verify path is within allowed directory
    if not str(full_path).startswith(str(UPLOAD_DIR)):
        raise ValueError("Path traversal detected")
    
    return full_path
```

## Unsafe Patterns (Flag These)

### Trusting User Input
```python
# DANGER: No validation
user_role = request.form['role']  # VULNERABLE
execute_as_role(user_role)

# DANGER: Denylist instead of allowlist
if '<script>' not in user_input:  # VULNERABLE - easy to bypass
    save_comment(user_input)
```

### Unsafe File Handling
```python
# DANGER: Using user filename directly
filename = request.files['file'].filename
path = f"/uploads/{filename}"  # VULNERABLE - path traversal

# DANGER: Trusting Content-Type
if file.content_type == 'image/jpeg':  # VULNERABLE - can be spoofed
    save_file(file)
```

### SSRF Vulnerabilities
```python
# DANGER: Fetching arbitrary URLs
url = request.form['url']
response = requests.get(url)  # VULNERABLE - SSRF

# DANGER: Allowing redirects
requests.get(url, allow_redirects=True)  # VULNERABLE - bypass protections
```

### XXE Vulnerabilities
```python
# DANGER: Default XML parser with entities enabled
from xml.etree.ElementTree import parse
tree = parse(user_xml_file)  # VULNERABLE to XXE
```

## Security Checklist

After implementing input handling:

- [ ] All input validated on server side
- [ ] Allowlist validation used (not denylist)
- [ ] Input length/size limits enforced
- [ ] File uploads validated by content, not just extension
- [ ] Uploaded files stored outside webroot
- [ ] Random filenames generated for uploads
- [ ] URLs validated against allowlist (SSRF prevention)
- [ ] XML parsing has external entities disabled
- [ ] Path traversal characters rejected
- [ ] Redirects validated or disabled

## Source References

- [Input Validation Cheat Sheet](../../security-guidelines/Input_Validation_Cheat_Sheet.md)
- [File Upload Cheat Sheet](../../security-guidelines/File_Upload_Cheat_Sheet.md)
- [Server Side Request Forgery Prevention Cheat Sheet](../../security-guidelines/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.md)
- [XML External Entity Prevention Cheat Sheet](../../security-guidelines/XML_External_Entity_Prevention_Cheat_Sheet.md)
- [XML Security Cheat Sheet](../../security-guidelines/XML_Security_Cheat_Sheet.md)
- [Deserialization Cheat Sheet](../../security-guidelines/Deserialization_Cheat_Sheet.md)
- [Unvalidated Redirects and Forwards Cheat Sheet](../../security-guidelines/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.md)

