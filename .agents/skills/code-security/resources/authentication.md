# Authentication & Authorization Security

## When This Applies

- Implementing user login/logout functionality
- Storing or validating passwords
- Managing user sessions
- Implementing access control (who can access what)
- Building OAuth2/OIDC integrations
- Implementing MFA/2FA
- Password reset flows
- API authentication

## Key Rules

### MUST Do

- Hash passwords with Argon2id, bcrypt, or scrypt (NEVER MD5, SHA1, or SHA256 alone)
- Use unique, random salts for each password (modern algorithms do this automatically)
- Enforce minimum password length: 8 chars with MFA, 15 chars without MFA
- Allow passwords up to at least 64 characters
- Use cryptographically secure session IDs with at least 64 bits of entropy
- Transmit credentials only over HTTPS/TLS
- Implement account lockout or rate limiting after failed attempts
- Validate permissions on EVERY request (not just at login)
- Use deny-by-default for authorization decisions
- Invalidate sessions on logout and password change
- Set secure cookie flags: `HttpOnly`, `Secure`, `SameSite`

### MUST NOT Do

- Store passwords in plaintext or with reversible encryption
- Use predictable session IDs or user-controlled session tokens
- Expose whether a username exists in error messages
- Allow unlimited login attempts without rate limiting
- Trust client-side authorization checks alone
- Use URL parameters for session IDs
- Store sensitive data in JWT payloads (they're base64, not encrypted)
- Skip re-authentication for sensitive operations

### SHOULD Do

- Implement MFA for sensitive applications
- Check passwords against breach databases (Have I Been Pwned)
- Use password strength meters (zxcvbn library)
- Implement session timeout and absolute timeout
- Rotate session IDs after authentication
- Log authentication events (without logging passwords)
- Prefer ABAC/ReBAC over RBAC for complex authorization

## Safe Patterns

### Python (Django)

```python
# SECURITY: Django's built-in auth handles password hashing with PBKDF2/Argon2
from django.contrib.auth.hashers import make_password, check_password

# Hash a password
hashed = make_password(raw_password)

# Verify a password
is_valid = check_password(raw_password, hashed)

# SECURITY: Use Django's authentication decorators
from django.contrib.auth.decorators import login_required, permission_required

@login_required
@permission_required('app.view_sensitive_data', raise_exception=True)
def sensitive_view(request):
    # SECURITY: Check object-level permissions
    obj = get_object_or_404(Resource, pk=pk)
    if obj.owner != request.user:
        raise PermissionDenied
    return render(request, 'template.html', {'obj': obj})
```

### Python (Flask)

```python
# SECURITY: Use werkzeug's secure password hashing
from werkzeug.security import generate_password_hash, check_password_hash

# Hash password with scrypt (default) or pbkdf2
hashed = generate_password_hash(password, method='scrypt')

# Verify password - uses constant-time comparison
is_valid = check_password_hash(hashed, password)

# SECURITY: Session configuration
app.config.update(
    SESSION_COOKIE_SECURE=True,      # HTTPS only
    SESSION_COOKIE_HTTPONLY=True,    # No JavaScript access
    SESSION_COOKIE_SAMESITE='Lax',   # CSRF protection
    PERMANENT_SESSION_LIFETIME=3600   # 1 hour timeout
)
```

### Go

```go
import "golang.org/x/crypto/bcrypt"

// SECURITY: Hash password with bcrypt (cost factor 10+)
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

// SECURITY: Verify password with constant-time comparison
func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// SECURITY: Authorization middleware
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := getUserFromSession(r)
        if user == nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        // SECURITY: Check permissions on every request
        if !user.HasPermission(r.URL.Path, r.Method) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### JavaScript/Node.js

```javascript
import bcrypt from 'bcrypt';

// SECURITY: Hash password with bcrypt, cost factor 12
const SALT_ROUNDS = 12;

async function hashPassword(password) {
    // SECURITY: bcrypt automatically generates and stores salt
    return await bcrypt.hash(password, SALT_ROUNDS);
}

async function verifyPassword(password, hash) {
    // SECURITY: Constant-time comparison built into bcrypt
    return await bcrypt.compare(password, hash);
}

// SECURITY: Session configuration (Express)
app.use(session({
    secret: process.env.SESSION_SECRET,  // From environment, not hardcoded
    name: 'sessionId',                   // Generic name, not default
    resave: false,
    saveUninitialized: false,
    cookie: {
        secure: true,                    // HTTPS only
        httpOnly: true,                  // No JavaScript access
        sameSite: 'lax',                 // CSRF protection
        maxAge: 3600000                  // 1 hour
    }
}));

// SECURITY: Authorization middleware
function requirePermission(permission) {
    return (req, res, next) => {
        if (!req.user) {
            return res.status(401).json({ error: 'Unauthorized' });
        }
        // SECURITY: Check on every request, not just login
        if (!req.user.permissions.includes(permission)) {
            return res.status(403).json({ error: 'Forbidden' });
        }
        next();
    };
}
```

### PHP (Laravel)

```php
// SECURITY: Laravel uses bcrypt by default
use Illuminate\Support\Facades\Hash;

// Hash password
$hashed = Hash::make($password);

// Verify password
if (Hash::check($password, $hashed)) {
    // Valid
}

// SECURITY: Authorization using policies
class PostPolicy
{
    // SECURITY: Object-level authorization
    public function update(User $user, Post $post)
    {
        return $user->id === $post->user_id;
    }
}

// In controller
public function update(Request $request, Post $post)
{
    // SECURITY: Authorize on every request
    $this->authorize('update', $post);
    // ...
}
```

## Unsafe Patterns (Flag These)

### Weak Password Hashing
```python
# DANGER: MD5/SHA1/SHA256 without salt are NOT for passwords
import hashlib
hashed = hashlib.md5(password.encode()).hexdigest()  # VULNERABLE
hashed = hashlib.sha256(password.encode()).hexdigest()  # VULNERABLE
```

### Predictable Session IDs
```python
# DANGER: Sequential or predictable session IDs
session_id = str(user_id) + "_" + str(int(time.time()))  # VULNERABLE
```

### Missing Authorization Checks
```javascript
// DANGER: No permission check on object access
app.get('/api/documents/:id', async (req, res) => {
    // VULNERABLE: Anyone can access any document by ID
    const doc = await Document.findById(req.params.id);
    res.json(doc);
});
```

### Information Disclosure in Auth Errors
```python
# DANGER: Reveals whether username exists
if not user_exists(username):
    return "Username not found"  # VULNERABLE
elif not check_password(password, user.password):
    return "Incorrect password"  # VULNERABLE

# SAFE: Generic error message
return "Invalid username or password"
```

### Hardcoded Secrets
```javascript
// DANGER: Session secret in source code
app.use(session({ secret: 'my-secret-key' }));  // VULNERABLE

// SAFE: Use environment variables
app.use(session({ secret: process.env.SESSION_SECRET }));
```

## Security Checklist

After implementing authentication/authorization:

- [ ] Passwords hashed with Argon2id, bcrypt, or scrypt
- [ ] Password requirements enforced (length, not complexity rules)
- [ ] Session IDs are random and unpredictable
- [ ] Cookies have Secure, HttpOnly, SameSite flags
- [ ] Authorization checked on every request
- [ ] Object-level permissions validated (not just role-based)
- [ ] Sensitive operations require re-authentication
- [ ] Failed login attempts are rate-limited
- [ ] Sessions invalidated on logout/password change
- [ ] Generic error messages (don't leak username existence)
- [ ] No secrets hardcoded in source code

## Source References

- [Authentication Cheat Sheet](../../security-guidelines/Authentication_Cheat_Sheet.md)
- [Password Storage Cheat Sheet](../../security-guidelines/Password_Storage_Cheat_Sheet.md)
- [Session Management Cheat Sheet](../../security-guidelines/Session_Management_Cheat_Sheet.md)
- [Authorization Cheat Sheet](../../security-guidelines/Authorization_Cheat_Sheet.md)
- [Multifactor Authentication Cheat Sheet](../../security-guidelines/Multifactor_Authentication_Cheat_Sheet.md)
- [Forgot Password Cheat Sheet](../../security-guidelines/Forgot_Password_Cheat_Sheet.md)
- [Credential Stuffing Prevention Cheat Sheet](../../security-guidelines/Credential_Stuffing_Prevention_Cheat_Sheet.md)
- [OAuth2 Cheat Sheet](../../security-guidelines/OAuth2_Cheat_Sheet.md)
- [Mass Assignment Cheat Sheet](../../security-guidelines/Mass_Assignment_Cheat_Sheet.md)
- [Insecure Direct Object Reference Prevention Cheat Sheet](../../security-guidelines/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.md)

