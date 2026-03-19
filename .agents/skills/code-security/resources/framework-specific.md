# Framework-Specific Security

## When This Applies

- Developing with Node.js/Express
- Building Django applications
- Creating Laravel/PHP applications
- Developing Ruby on Rails apps
- Configuring PHP environments

## Node.js / Express Security

### Key Rules

- Never use `eval()` or `Function()` constructor with user input
- Avoid `child_process.exec()` with user input; use `execFile()` or `spawn()`
- Use `--ignore-scripts` when installing npm packages
- Enable strict mode (`'use strict'`) in all files
- Use security linters (eslint-plugin-security)
- Set appropriate HTTP headers with `helmet`
- Validate all input with libraries like `joi` or `zod`

### Safe Patterns

```javascript
'use strict';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import { z } from 'zod';

const app = express();

// SECURITY: Apply security headers
app.use(helmet());

// SECURITY: Rate limiting
app.use('/api/', rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 100
}));

// SECURITY: Disable x-powered-by header
app.disable('x-powered-by');

// SECURITY: Parse JSON with size limit
app.use(express.json({ limit: '10kb' }));

// SECURITY: Input validation with Zod
const UserSchema = z.object({
    email: z.string().email(),
    password: z.string().min(8),
});

app.post('/register', (req, res) => {
    const result = UserSchema.safeParse(req.body);
    if (!result.success) {
        return res.status(400).json({ error: 'Invalid input' });
    }
    // Process validated data
});

// SECURITY: Secure session configuration
app.use(session({
    secret: process.env.SESSION_SECRET,
    name: 'sessionId',
    cookie: {
        secure: true,
        httpOnly: true,
        sameSite: 'strict',
        maxAge: 3600000
    },
    resave: false,
    saveUninitialized: false
}));

// SECURITY: Prevent prototype pollution
app.use(express.json({ 
    reviver: (key, value) => {
        if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
            return undefined;
        }
        return value;
    }
}));
```

### Unsafe Patterns to Avoid

```javascript
// DANGER: eval with user input
eval(userInput);  // VULNERABLE

// DANGER: Function constructor
new Function(userInput);  // VULNERABLE

// DANGER: exec with user input (shell injection)
child_process.exec(`ls ${userInput}`);  // VULNERABLE

// SAFE: Use execFile with arguments array and -- to prevent option injection
child_process.execFile('ls', ['--', userInput]);

// DANGER: Regex DoS
const regex = new RegExp(userInput);  // VULNERABLE if malicious pattern

// DANGER: Unvalidated redirects
res.redirect(req.query.url);  // VULNERABLE
```

---

## Django Security

### Key Rules

- Never set `DEBUG = True` in production
- Use Django's built-in CSRF protection
- Use `@login_required` for protected views
- Use Django ORM (avoids SQL injection by default)
- Configure `ALLOWED_HOSTS` properly
- Use Django's password validators
- Set secure cookie settings

### Safe Patterns

```python
# settings.py

# SECURITY: Never debug in production
DEBUG = False

# SECURITY: Restrict allowed hosts
ALLOWED_HOSTS = ['example.com', 'www.example.com']

# SECURITY: HTTPS settings
SECURE_SSL_REDIRECT = True
SECURE_HSTS_SECONDS = 31536000
SECURE_HSTS_INCLUDE_SUBDOMAINS = True
SECURE_HSTS_PRELOAD = True

# SECURITY: Cookie settings
SESSION_COOKIE_SECURE = True
SESSION_COOKIE_HTTPONLY = True
SESSION_COOKIE_SAMESITE = 'Lax'
CSRF_COOKIE_SECURE = True
CSRF_COOKIE_HTTPONLY = True

# SECURITY: Content security
SECURE_CONTENT_TYPE_NOSNIFF = True
X_FRAME_OPTIONS = 'DENY'
SECURE_BROWSER_XSS_FILTER = True

# SECURITY: Password validators
AUTH_PASSWORD_VALIDATORS = [
    {'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator'},
    {'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
     'OPTIONS': {'min_length': 12}},
    {'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator'},
    {'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator'},
]
```

```python
# views.py
from django.contrib.auth.decorators import login_required, permission_required
from django.views.decorators.http import require_http_methods
from django.core.exceptions import PermissionDenied

# SECURITY: Require authentication
@login_required
def protected_view(request):
    return render(request, 'protected.html')

# SECURITY: Require specific permission
@permission_required('app.can_edit', raise_exception=True)
def edit_view(request, pk):
    obj = get_object_or_404(MyModel, pk=pk)
    # SECURITY: Check object-level permission
    if obj.owner != request.user:
        raise PermissionDenied
    return render(request, 'edit.html', {'obj': obj})

# SECURITY: Restrict HTTP methods
@require_http_methods(["GET", "POST"])
def my_view(request):
    pass

# SECURITY: Django ORM is safe by default
def search(request):
    query = request.GET.get('q', '')
    # SAFE: ORM prevents SQL injection
    results = MyModel.objects.filter(name__icontains=query)
    return render(request, 'results.html', {'results': results})
```

### Unsafe Patterns to Avoid

```python
# DANGER: Raw SQL with user input
MyModel.objects.raw(f"SELECT * FROM mymodel WHERE name = '{user_input}'")  # VULNERABLE

# SAFE: Use parameters
MyModel.objects.raw("SELECT * FROM mymodel WHERE name = %s", [user_input])

# DANGER: Using |safe filter with user content
{{ user_content|safe }}  # VULNERABLE unless sanitized

# DANGER: DEBUG in production
DEBUG = True  # VULNERABLE

# DANGER: Wildcard allowed hosts
ALLOWED_HOSTS = ['*']  # VULNERABLE
```

---

## Laravel / PHP Security

### Key Rules

- Always use Eloquent ORM or query builder (prevents SQL injection)
- Use Laravel's CSRF protection (enabled by default)
- Validate all input with Form Requests
- Use `bcrypt()` or `Hash::make()` for passwords
- Configure `APP_DEBUG=false` in production
- Use Laravel's built-in authentication scaffolding
- Escape output with `{{ }}` (Blade auto-escapes)

### Safe Patterns

```php
// .env (production)
APP_DEBUG=false
APP_ENV=production
SESSION_SECURE_COOKIE=true

// config/session.php
'secure' => true,
'http_only' => true,
'same_site' => 'lax',
```

```php
<?php
// SECURITY: Form Request validation
namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class StoreUserRequest extends FormRequest
{
    public function rules()
    {
        return [
            'email' => 'required|email|unique:users',
            'password' => 'required|min:12|confirmed',
            'name' => 'required|string|max:255',
        ];
    }
}

// Controller
public function store(StoreUserRequest $request)
{
    // SECURITY: Validated data only
    $validated = $request->validated();
    
    // SECURITY: Hash password
    User::create([
        'email' => $validated['email'],
        'name' => $validated['name'],
        'password' => Hash::make($validated['password']),
    ]);
}

// SECURITY: Authorization with policies
public function update(Request $request, Post $post)
{
    $this->authorize('update', $post);  // Checks PostPolicy
    // ...
}

// SECURITY: Eloquent ORM prevents SQL injection
$users = User::where('email', $request->email)->get();  // SAFE
```

```blade
{{-- SECURITY: Blade auto-escapes output --}}
<p>{{ $user->name }}</p>  {{-- SAFE: Auto-escaped --}}

{{-- DANGER: Unescaped output --}}
<p>{!! $user->bio !!}</p>  {{-- VULNERABLE unless sanitized --}}
```

### Unsafe Patterns to Avoid

```php
// DANGER: Raw SQL with user input
DB::select("SELECT * FROM users WHERE email = '$email'");  // VULNERABLE

// SAFE: Use parameter binding
DB::select("SELECT * FROM users WHERE email = ?", [$email]);

// DANGER: Mass assignment without protection
User::create($request->all());  // VULNERABLE

// SAFE: Use validated data or fillable
User::create($request->validated());
```

---

## Ruby on Rails Security

### Key Rules

- Use strong parameters for mass assignment protection
- Rails escapes output by default in views
- Use `authenticate_user!` (Devise) for protected actions
- Configure `force_ssl` in production
- Use `has_secure_password` for password hashing
- Keep `config.consider_all_requests_local = false` in production
- Use parameterized queries (ActiveRecord does this by default)

### Safe Patterns

```ruby
# config/environments/production.rb
Rails.application.configure do
  # SECURITY: Force HTTPS
  config.force_ssl = true
  
  # SECURITY: Don't show detailed errors
  config.consider_all_requests_local = false
  
  # SECURITY: Secure cookies
  config.session_store :cookie_store,
    key: '_myapp_session',
    secure: true,
    httponly: true,
    same_site: :lax
end
```

```ruby
# SECURITY: Strong parameters
class UsersController < ApplicationController
  before_action :authenticate_user!
  
  def create
    @user = User.new(user_params)
    if @user.save
      redirect_to @user
    else
      render :new
    end
  end
  
  private
  
  # SECURITY: Whitelist allowed parameters
  def user_params
    params.require(:user).permit(:name, :email, :password, :password_confirmation)
  end
end

# SECURITY: Authorization with Pundit
class PostsController < ApplicationController
  def update
    @post = Post.find(params[:id])
    authorize @post  # SECURITY: Check policy
    # ...
  end
end

# SECURITY: ActiveRecord prevents SQL injection
User.where(email: params[:email])  # SAFE
User.where("email = ?", params[:email])  # SAFE
```

### Unsafe Patterns to Avoid

```ruby
# DANGER: String interpolation in queries
User.where("email = '#{params[:email]}'")  # VULNERABLE

# SAFE: Use parameterized queries
User.where("email = ?", params[:email])
User.where(email: params[:email])

# DANGER: render user-controlled content unescaped
<%= raw user_input %>  # VULNERABLE
<%= user_input.html_safe %>  # VULNERABLE

# DANGER: Open redirect
redirect_to params[:url]  # VULNERABLE

# SAFE: Validate redirect URL
redirect_to params[:url] if params[:url].start_with?('/')
```

---

## PHP Configuration Security

### Key Rules (php.ini)

```ini
; SECURITY: Disable dangerous functions
disable_functions = exec,passthru,shell_exec,system,proc_open,popen,curl_exec,curl_multi_exec,parse_ini_file,show_source

; SECURITY: Hide PHP version
expose_php = Off

; SECURITY: Session security
session.cookie_httponly = On
session.cookie_secure = On
session.use_strict_mode = On
session.cookie_samesite = Lax

; SECURITY: Limit file uploads
file_uploads = On
upload_max_filesize = 2M
max_file_uploads = 5

; SECURITY: Error handling
display_errors = Off
log_errors = On
error_reporting = E_ALL

; SECURITY: Disable remote file inclusion
allow_url_fopen = Off
allow_url_include = Off

; SECURITY: Open basedir restriction
open_basedir = /var/www/html:/tmp
```

### Safe PHP Patterns

```php
<?php
// SECURITY: Prepared statements with PDO
$stmt = $pdo->prepare('SELECT * FROM users WHERE email = :email');
$stmt->execute(['email' => $email]);
$user = $stmt->fetch();

// SECURITY: Escape output
echo htmlspecialchars($userInput, ENT_QUOTES, 'UTF-8');

// SECURITY: Password hashing
$hash = password_hash($password, PASSWORD_DEFAULT);
if (password_verify($inputPassword, $hash)) {
    // Valid
}

// SECURITY: CSRF token
session_start();
if (empty($_SESSION['csrf_token'])) {
    $_SESSION['csrf_token'] = bin2hex(random_bytes(32));
}

// In form
echo '<input type="hidden" name="csrf_token" value="' . $_SESSION['csrf_token'] . '">';

// Validate
if (!hash_equals($_SESSION['csrf_token'], $_POST['csrf_token'])) {
    die('CSRF validation failed');
}
```

## Security Checklist by Framework

### Node.js/Express
- [ ] Helmet middleware configured
- [ ] Rate limiting enabled
- [ ] Input validation with Zod/Joi
- [ ] Secure session configuration
- [ ] No eval() or Function() with user input

### Django
- [ ] DEBUG = False in production
- [ ] ALLOWED_HOSTS configured
- [ ] HTTPS settings enabled
- [ ] Secure cookie flags set
- [ ] CSRF protection active

### Laravel
- [ ] APP_DEBUG=false in production
- [ ] Form Request validation used
- [ ] Passwords hashed with Hash::make()
- [ ] Authorization policies defined
- [ ] Blade escaping ({{ }}) used

### Rails
- [ ] force_ssl enabled
- [ ] Strong parameters used
- [ ] Pundit/CanCanCan for authorization
- [ ] No raw SQL with user input
- [ ] consider_all_requests_local = false

### PHP
- [ ] Dangerous functions disabled
- [ ] expose_php = Off
- [ ] Secure session settings
- [ ] Prepared statements used
- [ ] display_errors = Off

## Source References

- [Node.js Security Cheat Sheet](../../security-guidelines/Nodejs_Security_Cheat_Sheet.md)
- [Django Security Cheat Sheet](../../security-guidelines/Django_Security_Cheat_Sheet.md)
- [Laravel Cheat Sheet](../../security-guidelines/Laravel_Cheat_Sheet.md)
- [Ruby on Rails Cheat Sheet](../../security-guidelines/Ruby_on_Rails_Cheat_Sheet.md)
- [PHP Configuration Cheat Sheet](../../security-guidelines/PHP_Configuration_Cheat_Sheet.md)

