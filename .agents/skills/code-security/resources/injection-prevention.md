# Injection Prevention Security

## When This Applies

- Writing database queries (SQL, NoSQL, ORM)
- Rendering user-provided content in HTML/JavaScript
- Executing system commands
- Processing forms that modify server state
- Building URLs with user input
- Handling any user input that will be interpreted by another system

## Key Rules

### MUST Do

- Use parameterized queries / prepared statements for ALL database operations
- Encode output based on context (HTML, JavaScript, URL, CSS)
- Use CSRF tokens for all state-changing operations
- Validate and sanitize all user input on the server side
- Use framework's built-in protections (auto-escaping templates, CSRF middleware)
- Set `Content-Type: application/json` for JSON responses (not `text/html`)
- Use `X-Frame-Options` or CSP `frame-ancestors` to prevent clickjacking

### MUST NOT Do

- Concatenate user input into SQL queries
- Use `eval()`, `exec()`, or dynamic code execution with user input
- Insert user content with `innerHTML`, `dangerouslySetInnerHTML`, or `v-html` without sanitization
- Trust client-side validation alone
- Use GET requests for state-changing operations
- Directly pass user input to system shell commands
- Use `document.write()` with user-controlled data

### SHOULD Do

- Use Content Security Policy (CSP) headers
- Implement defense in depth (multiple layers of protection)
- Use DOMPurify or similar for HTML sanitization when needed
- Prefer built-in library functions over shell commands
- Use the `--` argument terminator for command-line tools

## Safe Patterns

### SQL Injection Prevention

#### Python

```python
# SECURITY: Parameterized query prevents SQL injection
cursor.execute(
    "SELECT * FROM users WHERE email = %s AND status = %s",
    (email, status)
)

# SECURITY: Django ORM is safe by default
User.objects.filter(email=email, status=status)

# SECURITY: SQLAlchemy with parameters
session.query(User).filter(User.email == email).all()
```

#### Go

```go
// SECURITY: Parameterized query with placeholders
rows, err := db.Query(
    "SELECT * FROM users WHERE email = $1 AND status = $2",
    email, status,
)

// SECURITY: Using prepared statements
stmt, err := db.Prepare("SELECT * FROM users WHERE id = $1")
defer stmt.Close()
rows, err := stmt.Query(userID)
```

#### JavaScript/Node.js

```javascript
// SECURITY: Parameterized query with pg (PostgreSQL)
const result = await pool.query(
    'SELECT * FROM users WHERE email = $1 AND status = $2',
    [email, status]
);

// SECURITY: Prisma ORM is safe by default
const user = await prisma.user.findMany({
    where: { email, status }
});

// SECURITY: MongoDB with proper query structure
const user = await User.findOne({ email: email, status: status });

// DANGER: NoSQL injection with $where or $regex from user input
// Never do: User.find({ $where: userInput })
```

#### PHP

```php
// SECURITY: PDO with prepared statements
$stmt = $pdo->prepare('SELECT * FROM users WHERE email = :email');
$stmt->execute(['email' => $email]);

// SECURITY: Laravel Eloquent is safe by default
$users = User::where('email', $email)->get();

// SECURITY: Laravel query builder with bindings
$users = DB::select('SELECT * FROM users WHERE email = ?', [$email]);
```

### XSS Prevention

#### JavaScript/React

```jsx
// SECURITY: React auto-escapes by default - this is SAFE
function UserProfile({ user }) {
    return <div>{user.name}</div>;  // Auto-escaped
}

// DANGER: dangerouslySetInnerHTML bypasses escaping
// Only use with sanitized content:
import DOMPurify from 'dompurify';

function SafeHTML({ html }) {
    // SECURITY: Sanitize before using dangerouslySetInnerHTML
    return <div dangerouslySetInnerHTML={{ 
        __html: DOMPurify.sanitize(html) 
    }} />;
}

// SECURITY: Use textContent for DOM manipulation
element.textContent = userInput;  // SAFE - auto-escapes

// DANGER: innerHTML is vulnerable
// element.innerHTML = userInput;  // VULNERABLE
```

#### Python (Jinja2/Django)

```python
# SECURITY: Jinja2 auto-escapes by default
# {{ user_input }} is automatically escaped

# DANGER: |safe filter bypasses escaping
# {{ user_input|safe }}  # VULNERABLE - only use with trusted/sanitized content

# SECURITY: Django templates auto-escape by default
# {{ user_input }} is safe
```

#### Go

```go
import "html/template"

// SECURITY: html/template auto-escapes by default
tmpl := template.Must(template.New("page").Parse(`
    <div>{{.UserInput}}</div>
`))
// UserInput is automatically HTML-escaped

// DANGER: text/template does NOT escape
// import "text/template"  // VULNERABLE for HTML
```

### CSRF Prevention

#### Python (Django)

```python
# SECURITY: Django has built-in CSRF protection
# Ensure middleware is enabled:
MIDDLEWARE = [
    'django.middleware.csrf.CsrfViewMiddleware',
    # ...
]

# In templates, use the csrf_token tag:
# <form method="post">
#     {% csrf_token %}
#     ...
# </form>
```

#### JavaScript (Express)

```javascript
import csrf from 'csurf';

// SECURITY: CSRF middleware for Express
const csrfProtection = csrf({ cookie: true });

app.get('/form', csrfProtection, (req, res) => {
    res.render('form', { csrfToken: req.csrfToken() });
});

app.post('/submit', csrfProtection, (req, res) => {
    // Token is automatically validated
    // ...
});

// SECURITY: For SPAs, use custom header approach
// Frontend sends: X-CSRF-Token header
// Backend validates header matches session token
```

#### Cookie Configuration

```javascript
// SECURITY: SameSite cookies help prevent CSRF
res.cookie('session', token, {
    httpOnly: true,
    secure: true,
    sameSite: 'lax'  // or 'strict' for more protection
});
```

### Command Injection Prevention

#### Python

```python
import subprocess
import shlex

# SECURITY: Use subprocess with list arguments (no shell)
result = subprocess.run(
    ['ls', '-la', directory],  # Arguments as list
    capture_output=True,
    shell=False  # IMPORTANT: Never use shell=True with user input
)

# SECURITY: If shell is needed, use shlex.quote
safe_filename = shlex.quote(user_filename)

# BETTER: Use built-in functions instead of shell commands
import os
files = os.listdir(directory)  # Instead of subprocess(['ls', directory])
```

#### Go

```go
import "os/exec"

// SECURITY: Pass arguments separately, not as shell string
cmd := exec.Command("grep", pattern, filename)
output, err := cmd.Output()

// DANGER: Never use shell execution with user input
// cmd := exec.Command("sh", "-c", "grep " + userInput)  // VULNERABLE
```

#### JavaScript/Node.js

```javascript
import { execFile, spawn } from 'child_process';

// SECURITY: execFile with arguments array (no shell)
execFile('ls', ['-la', directory], (error, stdout) => {
    console.log(stdout);
});

// SECURITY: spawn with arguments array
const child = spawn('grep', [pattern, filename]);

// DANGER: exec() uses shell - vulnerable to injection
// exec(`ls -la ${userInput}`);  // VULNERABLE

// BETTER: Use built-in fs functions
import fs from 'fs';
const files = fs.readdirSync(directory);
```

#### PHP

```php
// SECURITY: escapeshellarg for single arguments
$safe_arg = escapeshellarg($user_input);
$output = shell_exec("command " . $safe_arg);

// SECURITY: escapeshellcmd for entire command
$safe_cmd = escapeshellcmd($command);

// BETTER: Use built-in PHP functions
$files = scandir($directory);  // Instead of shell_exec("ls $directory")
```

## Unsafe Patterns (Flag These)

### SQL Injection Vulnerabilities
```python
# DANGER: String concatenation in SQL
query = "SELECT * FROM users WHERE email = '" + email + "'"
cursor.execute(query)  # VULNERABLE

# DANGER: f-strings in SQL
cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")  # VULNERABLE

# DANGER: .format() in SQL
query = "SELECT * FROM users WHERE name = '{}'".format(name)  # VULNERABLE
```

### XSS Vulnerabilities
```javascript
// DANGER: innerHTML with user content
element.innerHTML = userInput;  // VULNERABLE

// DANGER: document.write with user content
document.write(userInput);  // VULNERABLE

// DANGER: React without sanitization
<div dangerouslySetInnerHTML={{ __html: userInput }} />  // VULNERABLE

// DANGER: jQuery html() with user content
$('#element').html(userInput);  // VULNERABLE
```

### Command Injection Vulnerabilities
```python
# DANGER: shell=True with user input
subprocess.run(f"echo {user_input}", shell=True)  # VULNERABLE

# DANGER: os.system with user input
os.system("cat " + filename)  # VULNERABLE
```

### CSRF Vulnerabilities
```html
<!-- DANGER: Form without CSRF token -->
<form method="POST" action="/transfer">
    <input name="amount" value="1000">
    <button>Transfer</button>
</form>  <!-- VULNERABLE -->

<!-- DANGER: State-changing GET request -->
<a href="/delete?id=123">Delete</a>  <!-- VULNERABLE -->
```

## Security Checklist

After implementing features with user input:

- [ ] Database queries use parameterized statements
- [ ] User content is output-encoded for the context (HTML, JS, URL)
- [ ] CSRF tokens included in all state-changing forms
- [ ] No user input passed to eval(), exec(), or shell commands
- [ ] Content Security Policy headers configured
- [ ] X-Frame-Options or frame-ancestors set for clickjacking protection
- [ ] Input validation applied on server side
- [ ] JSON responses use Content-Type: application/json

## Source References

- [SQL Injection Prevention Cheat Sheet](../../security-guidelines/SQL_Injection_Prevention_Cheat_Sheet.md)
- [Cross Site Scripting Prevention Cheat Sheet](../../security-guidelines/Cross_Site_Scripting_Prevention_Cheat_Sheet.md)
- [DOM based XSS Prevention Cheat Sheet](../../security-guidelines/DOM_based_XSS_Prevention_Cheat_Sheet.md)
- [Cross-Site Request Forgery Prevention Cheat Sheet](../../security-guidelines/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.md)
- [OS Command Injection Defense Cheat Sheet](../../security-guidelines/OS_Command_Injection_Defense_Cheat_Sheet.md)
- [Query Parameterization Cheat Sheet](../../security-guidelines/Query_Parameterization_Cheat_Sheet.md)
- [Injection Prevention Cheat Sheet](../../security-guidelines/Injection_Prevention_Cheat_Sheet.md)
- [NoSQL Security Cheat Sheet](../../security-guidelines/NoSQL_Security_Cheat_Sheet.md)
- [Clickjacking Defense Cheat Sheet](../../security-guidelines/Clickjacking_Defense_Cheat_Sheet.md)
- [Prototype Pollution Prevention Cheat Sheet](../../security-guidelines/Prototype_Pollution_Prevention_Cheat_Sheet.md)

