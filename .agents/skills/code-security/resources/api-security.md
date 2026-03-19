# API Security

## When This Applies

- Building REST APIs
- Implementing GraphQL endpoints
- Creating gRPC services
- Using WebSocket connections
- Designing microservices architecture
- Integrating with third-party APIs
- Implementing API authentication (JWT, API keys, OAuth)

## Key Rules

### MUST Do

- Use HTTPS for all API endpoints (no exceptions)
- Validate and verify JWT tokens (if used) on every request (signature, issuer, audience, expiration)
- Implement rate limiting and throttling
- Apply authorization checks at every endpoint
- Validate all input data against strict schemas
- Return appropriate HTTP status codes
- Use parameterized queries in resolvers (GraphQL, REST)
- Set proper CORS policies (not `*` for authenticated APIs)
- Log API access and security events

### MUST NOT Do

- Accept unsigned JWTs (`{"alg":"none"}`)
- Trust the JWT header to select verification algorithm
- Expose sensitive data in error messages
- Allow unlimited query depth or complexity (GraphQL)
- Enable introspection in production (GraphQL)
- Use GET requests for state-changing operations
- Store API keys or secrets in client-side code
- Allow unrestricted CORS origins for authenticated endpoints

### SHOULD Do

- Use short-lived access tokens with refresh tokens
- Implement request signing for sensitive operations
- Add query cost analysis (GraphQL)
- Use pagination for list endpoints
- Implement idempotency keys for critical operations
- Version your APIs
- Add request/response logging (without sensitive data)

## Safe Patterns

### JWT Validation

#### Python

```python
import jwt
from datetime import datetime, timezone

# SECURITY: Always verify JWT with explicit algorithm
def validate_jwt(token: str, secret: str) -> dict:
    try:
        # SECURITY: Specify allowed algorithms explicitly
        payload = jwt.decode(
            token,
            secret,
            algorithms=["HS256"],  # Explicit algorithm - never trust header
            options={
                "require": ["exp", "iat", "iss", "aud"],
                "verify_exp": True,
                "verify_iss": True,
                "verify_aud": True,
            },
            issuer="your-app",
            audience="your-api"
        )
        return payload
    except jwt.ExpiredSignatureError:
        raise AuthError("Token has expired")
    except jwt.InvalidTokenError as e:
        raise AuthError(f"Invalid token: {e}")
```

#### Go

```go
import "github.com/golang-jwt/jwt/v5"

// SECURITY: Validate JWT with explicit algorithm and claims
func ValidateJWT(tokenString string, secret []byte) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            // SECURITY: Verify signing method explicitly
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return secret, nil
        },
        // SECURITY: Require standard claims
        jwt.WithValidMethods([]string{"HS256"}),
        jwt.WithIssuer("your-app"),
        jwt.WithAudience("your-api"),
    )
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}
```

#### JavaScript/Node.js

```javascript
import jwt from 'jsonwebtoken';

// SECURITY: Validate JWT with explicit options
function validateJWT(token, secret) {
    try {
        // SECURITY: Specify algorithms to prevent algorithm confusion
        const payload = jwt.verify(token, secret, {
            algorithms: ['HS256'],  // Explicit - don't trust header
            issuer: 'your-app',
            audience: 'your-api',
            complete: false
        });
        return payload;
    } catch (error) {
        if (error.name === 'TokenExpiredError') {
            throw new AuthError('Token has expired');
        }
        throw new AuthError('Invalid token');
    }
}

// SECURITY: JWT middleware for Express
function jwtMiddleware(req, res, next) {
    const authHeader = req.headers.authorization;
    if (!authHeader?.startsWith('Bearer ')) {
        return res.status(401).json({ error: 'Missing token' });
    }
    
    const token = authHeader.slice(7);
    try {
        req.user = validateJWT(token, process.env.JWT_SECRET);
        next();
    } catch (error) {
        return res.status(401).json({ error: error.message });
    }
}
```

### Rate Limiting

#### Python (FastAPI)

```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@app.get("/api/resource")
@limiter.limit("100/minute")  # SECURITY: Rate limit per IP
async def get_resource(request: Request):
    return {"data": "..."}

# SECURITY: Stricter limits for auth endpoints
@app.post("/api/login")
@limiter.limit("5/minute")  # Prevent brute force
async def login(request: Request, credentials: LoginRequest):
    # ...
```

#### JavaScript/Node.js (Express)

```javascript
import rateLimit from 'express-rate-limit';

// SECURITY: General rate limiter
const generalLimiter = rateLimit({
    windowMs: 60 * 1000,  // 1 minute
    max: 100,              // 100 requests per minute
    message: { error: 'Too many requests' },
    standardHeaders: true,
    legacyHeaders: false,
});

// SECURITY: Strict limiter for auth endpoints
const authLimiter = rateLimit({
    windowMs: 60 * 1000,
    max: 5,  // Only 5 attempts per minute
    message: { error: 'Too many login attempts' },
});

app.use('/api/', generalLimiter);
app.use('/api/auth/', authLimiter);
```

### GraphQL Security

```javascript
import { ApolloServer } from '@apollo/server';
import depthLimit from 'graphql-depth-limit';
import { createComplexityLimitRule } from 'graphql-validation-complexity';

const server = new ApolloServer({
    typeDefs,
    resolvers,
    validationRules: [
        // SECURITY: Limit query depth to prevent DoS
        depthLimit(5),
        // SECURITY: Limit query complexity
        createComplexityLimitRule(1000),
    ],
    // SECURITY: Disable introspection in production
    introspection: process.env.NODE_ENV !== 'production',
    plugins: [
        // SECURITY: Disable GraphQL Playground in production
        process.env.NODE_ENV === 'production'
            ? ApolloServerPluginLandingPageDisabled()
            : ApolloServerPluginLandingPageGraphQLPlayground(),
    ],
});

// SECURITY: Authorization in resolvers
const resolvers = {
    Query: {
        user: async (parent, { id }, context) => {
            // SECURITY: Check authorization
            if (!context.user) {
                throw new AuthenticationError('Must be logged in');
            }
            // SECURITY: Check object-level permissions
            if (context.user.id !== id && !context.user.isAdmin) {
                throw new ForbiddenError('Cannot access this user');
            }
            return await User.findById(id);
        },
    },
};
```

### CORS Configuration

#### Python (FastAPI)

```python
from fastapi.middleware.cors import CORSMiddleware

# SECURITY: Explicit allowed origins (never use "*" for authenticated APIs)
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "https://app.example.com",
        "https://admin.example.com",
    ],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Authorization", "Content-Type"],
)
```

#### JavaScript/Node.js (Express)

```javascript
import cors from 'cors';

// SECURITY: Explicit CORS configuration
const corsOptions = {
    origin: [
        'https://app.example.com',
        'https://admin.example.com',
    ],
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE'],
    allowedHeaders: ['Authorization', 'Content-Type'],
};

app.use(cors(corsOptions));

// DANGER: Never do this for authenticated APIs
// app.use(cors({ origin: '*' }));  // VULNERABLE
```

### API Key Validation

```javascript
// SECURITY: API key middleware
function apiKeyMiddleware(req, res, next) {
    const apiKey = req.headers['x-api-key'];
    
    if (!apiKey) {
        return res.status(401).json({ error: 'API key required' });
    }
    
    const validKey = process.env.API_KEY;
    
    // SECURITY: timingSafeEqual throws if lengths differ, so check first
    // Length check leaks timing info but prevents crashes
    const apiKeyBuffer = Buffer.from(apiKey);
    const validKeyBuffer = Buffer.from(validKey);
    
    if (apiKeyBuffer.length !== validKeyBuffer.length ||
        !crypto.timingSafeEqual(apiKeyBuffer, validKeyBuffer)) {
        // SECURITY: Don't reveal if key format is wrong vs invalid
        return res.status(401).json({ error: 'Invalid API key' });
    }
    
    next();
}
```

### WebSocket Security

```javascript
import { WebSocketServer } from 'ws';

const wss = new WebSocketServer({ server });

wss.on('connection', (ws, req) => {
    // SECURITY: Validate origin
    const origin = req.headers.origin;
    if (!allowedOrigins.includes(origin)) {
        ws.close(1008, 'Origin not allowed');
        return;
    }
    
    // SECURITY: Validate authentication token
    const token = new URL(req.url, 'ws://localhost').searchParams.get('token');
    try {
        const user = validateJWT(token, process.env.JWT_SECRET);
        ws.user = user;
    } catch (error) {
        ws.close(1008, 'Authentication failed');
        return;
    }
    
    ws.on('message', (data) => {
        // SECURITY: Validate and sanitize all incoming messages
        try {
            const message = JSON.parse(data);
            // Validate message schema...
        } catch (error) {
            ws.send(JSON.stringify({ error: 'Invalid message format' }));
        }
    });
});
```

## Unsafe Patterns (Flag These)

### JWT Vulnerabilities
```javascript
// DANGER: Accepting any algorithm (algorithm confusion attack)
jwt.verify(token, secret);  // VULNERABLE - no algorithm specified

// DANGER: Using "none" algorithm
const token = jwt.sign(payload, '', { algorithm: 'none' });  // VULNERABLE
```

### Missing Authorization
```javascript
// DANGER: No authorization check
app.get('/api/users/:id', async (req, res) => {
    const user = await User.findById(req.params.id);
    res.json(user);  // VULNERABLE - anyone can access any user
});
```

### Open CORS
```javascript
// DANGER: Wildcard CORS with credentials
app.use(cors({ 
    origin: '*', 
    credentials: true  // VULNERABLE
}));
```

### GraphQL Issues
```javascript
// DANGER: No depth limiting
const server = new ApolloServer({
    typeDefs,
    resolvers,
    introspection: true,  // VULNERABLE in production
    // No depth or complexity limits
});
```

## Security Checklist

After implementing an API:

- [ ] HTTPS enforced for all endpoints
- [ ] JWT validation includes algorithm, issuer, audience, expiration
- [ ] Rate limiting implemented (stricter for auth endpoints)
- [ ] Authorization checked at every endpoint
- [ ] Input validation with strict schemas
- [ ] CORS configured with explicit origins (not `*`)
- [ ] GraphQL introspection disabled in production
- [ ] GraphQL depth and complexity limits set
- [ ] Sensitive data not exposed in errors
- [ ] API access logged for security monitoring

## Source References

- [REST Security Cheat Sheet](../../security-guidelines/REST_Security_Cheat_Sheet.md)
- [GraphQL Cheat Sheet](../../security-guidelines/GraphQL_Cheat_Sheet.md)
- [gRPC Security Cheat Sheet](../../security-guidelines/gRPC_Security_Cheat_Sheet.md)
- [WebSocket Security Cheat Sheet](../../security-guidelines/WebSocket_Security_Cheat_Sheet.md)
- [Microservices Security Cheat Sheet](../../security-guidelines/Microservices_Security_Cheat_Sheet.md)
- [REST Assessment Cheat Sheet](../../security-guidelines/REST_Assessment_Cheat_Sheet.md)

