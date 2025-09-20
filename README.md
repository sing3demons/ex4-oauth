# OAuth2 Authorization Server

แอปพลิเคชัน backend OAuth2 Authorization Server ที่สมบูรณ์ พร้อมการรองรับ:
- 📝 User Registration และ Login
- 🔐 JWT Authentication (RS256 + HMAC256)
- 🛡️ OAuth2 Authorization Server
- 🆔 OpenID Connect (OIDC) กับ ID Token
- 🔒 PKCE (Proof Key for Code Exchange)
- 👤 Profile Management
- ♻️ Token Refresh
- ⚙️ OAuth2 Client Management
- 🔑 JWK (JSON Web Key) Support
- 🛡️ Advanced Rate Limiting
- 📊 Monitoring & Analytics

## คุณสมบัติหลัก

### Authentication & Authorization
- สมัครสมาชิกและเข้าสู่ระบบด้วย email/password
- JWT access tokens และ refresh tokens (RS256/HMAC256)
- OAuth2 Authorization Code Grant พร้อม PKCE
- OpenID Connect สำหรับ SSO พร้อม ID Token
- Consent screen สำหรับการขออนุญาต
- Rate limiting ป้องกัน brute force และ API abuse

### OAuth2 Server Endpoints
- `/api/auth/oauth/authorize` - Authorization endpoint
- `/api/auth/oauth/token` - Token endpoint  
- `/api/auth/oauth/userinfo` - UserInfo endpoint
- `/api/auth/oauth/.well-known/openid_configuration` - Discovery endpoint
- `/api/auth/oauth/.well-known/jwks.json` - JSON Web Key Set endpoint

### Security Features
- 🔐 RSA256 JWT signing พร้อม JWK support
- 🛡️ Token bucket rate limiting algorithm
- 🚫 Brute force protection (5 login attempts/minute)
- 🔒 OAuth2 rate limiting (20 requests/minute)
- 🌐 API rate limiting (100 requests/minute)
- 📊 Real-time rate limit monitoring
- 🧹 Automatic memory cleanup

### Client Management
- สร้าง OAuth2 clients
- จัดการ redirect URIs และ scopes
- Client credentials management
- Support multiple grant types

## การติดตั้งและรัน

### 1. ติดตั้ง Dependencies

```bash
go mod tidy
```

### 2. ตั้งค่า Environment Variables

สร้างไฟล์ `.env`:

```bash
# Server settings
PORT=8080
HOST=localhost
ENV=development

# Database
DATABASE_URL=auth.db

# JWT settings  
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h

# Security
BCRYPT_COST=12
RATE_LIMIT_PER_MIN=60

# CORS
ALLOWED_ORIGIN_1=http://localhost:3000
ALLOWED_ORIGIN_2=http://localhost:3001
```

### 3. รันแอปพลิเคชัน

```bash
go run main.go
```

Server จะรันที่ `http://localhost:8080`

## API Endpoints

### System
- `GET /health` - Health check
- `GET /api/info` - API information
- `GET /api/monitoring/rate-limits` - Rate limiting status

### Authentication
- `POST /api/auth/register` - สมัครสมาชิก
- `POST /api/auth/login` - เข้าสู่ระบบ (Rate Limited: 5/min)
- `POST /api/auth/refresh` - refresh token
- `POST /api/auth/logout` - ออกจากระบบ
- `GET /api/auth/profile` - ดู profile
- `PUT /api/auth/profile` - แก้ไข profile
- `POST /api/auth/change-password` - เปลี่ยนรหัสผ่าน

### OAuth2 Authorization Server (Rate Limited: 20/min)
- `GET /api/auth/oauth/authorize` - Authorization endpoint
- `POST /api/auth/oauth/token` - Token endpoint
- `POST /api/auth/oauth/consent` - Consent endpoint
- `GET /api/auth/oauth/userinfo` - UserInfo endpoint
- `GET /api/auth/oauth/.well-known/openid_configuration` - OIDC Discovery
- `GET /api/auth/oauth/.well-known/jwks.json` - JSON Web Key Set

### OAuth2 Client Management
- `POST /api/oauth2/clients` - สร้าง OAuth2 client
- `GET /api/oauth2/clients` - ดูรายการ clients
- `GET /api/oauth2/clients/:id` - ดู client ตาม ID
- `PUT /api/oauth2/clients/:id` - แก้ไข client
- `DELETE /api/oauth2/clients/:id` - ลบ client

### User Management
- `GET /api/users` - ดูรายการผู้ใช้

## ตัวอย่างการใช้งาน

### 1. สมัครสมาชิก

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### 2. เข้าสู่ระบบ

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 3. สร้าง OAuth2 Client

```bash
curl -X POST http://localhost:8080/api/oauth2/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "My App",
    "description": "My awesome application",
    "redirect_uris": ["http://localhost:3000/callback"],
    "scopes": ["openid", "email", "profile"],
    "grant_types": ["authorization_code", "refresh_token"]
  }'
```

### 4. OAuth2 Authorization Flow

#### ขั้นตอนที่ 1: Authorization Request

```
GET /api/auth/oauth/authorize?
  response_type=code&
  client_id=YOUR_CLIENT_ID&
  redirect_uri=http://localhost:3000/callback&
  scope=openid email profile&
  state=random_state&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256&
  nonce=random_nonce
```

#### ขั้นตอนที่ 2: Token Exchange

```bash
curl -X POST http://localhost:8080/api/auth/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d 'grant_type=authorization_code&
      code=AUTHORIZATION_CODE&
      redirect_uri=http://localhost:3000/callback&
      client_id=YOUR_CLIENT_ID&
      client_secret=YOUR_CLIENT_SECRET&
      code_verifier=CODE_VERIFIER'
```

**Response พร้อม ID Token:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "random-refresh-token",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "scope": "openid email profile"
}
```

#### ขั้นตอนที่ 3: UserInfo Request

```bash
curl -X GET http://localhost:8080/api/auth/oauth/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### 5. ID Token (OIDC)

ID Token จะถูกสร้างเมื่อมี `openid` scope และประกอบด้วย standard OIDC claims:

```json
{
  "sub": "1",
  "aud": "client_id",
  "iss": "ex4-oauth2",
  "iat": 1758374582,
  "exp": 1758375482,
  "auth_time": 1758374582,
  "nonce": "random_nonce",
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe",
  "preferred_username": "johndoe",
  "email": "john@example.com",
  "email_verified": true,
  "picture": "https://example.com/avatar.jpg"
}
```

### 6. JWK (JSON Web Key)

ดึง public keys สำหรับ verify JWT signatures:

```bash
curl -s http://localhost:8080/api/auth/oauth/.well-known/jwks.json
```

**Response:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-id-123",
      "n": "modulus-base64url",
      "e": "AQAB"
    }
  ]
}
```

### 7. Rate Limiting

ระบบมี rate limiting ป้องกัน abuse:

```bash
# ถ้าเกิน rate limit จะได้ response:
{
  "error": "rate_limit_exceeded",
  "error_description": "Too many requests. Please try again later.",
  "retry_after": 60
}
```

**Rate Limit Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 50
X-RateLimit-Reset: 1758374582
```

### 8. Monitoring Rate Limits

```bash
curl -s http://localhost:8080/api/monitoring/rate-limits
```

**Response:**
```json
{
  "status": "ok",
  "rate_limits": {
    "active_limiters": 3,
    "clients": {
      "localhost": {
        "tokens": 95,
        "max_tokens": 100,
        "last_refill": "2025-09-20T20:30:00Z"
      }
    }
  },
  "timestamp": 1758374582
}
```

## PKCE Implementation

แอปนี้รองรับ PKCE (Proof Key for Code Exchange) สำหรับความปลอดภัยเพิ่มเติม:

1. สร้าง `code_verifier` (43-128 characters)
2. สร้าง `code_challenge` จาก SHA256(code_verifier) และ base64url encode
3. ส่ง `code_challenge` และ `code_challenge_method=S256` ใน authorization request
4. ส่ง `code_verifier` ใน token request

## OpenID Connect (OIDC)

รองรับ OIDC discovery endpoint และ ID Token:

### Discovery Endpoint
```bash
curl http://localhost:8080/api/auth/oauth/.well-known/openid_configuration
```

### JWK Set Endpoint
```bash
curl http://localhost:8080/api/auth/oauth/.well-known/jwks.json
```

### ID Token Claims
- **Standard Claims**: sub, aud, iss, iat, exp, auth_time, nonce
- **Profile Claims**: name, given_name, family_name, preferred_username, picture
- **Email Claims**: email, email_verified
- **Address Claims**: ตาม OIDC specification

## Security Features

### JWT Security
- **RSA256 Algorithm**: มากกว่า HMAC256 สำหรับ OIDC compliance
- **Key ID (kid)**: ระบุ key ที่ใช้ sign ใน JWT header
- **JWK Support**: Public key distribution สำหรับ signature verification
- **Backward Compatible**: รองรับทั้ง RS256 และ HMAC256

### Rate Limiting
- **Token Bucket Algorithm**: Advanced rate limiting strategy
- **Per-IP Limiting**: แยก rate limit ตาม IP address
- **Graduated Limits**:
  - Login: 5 attempts/minute (ป้องกัน brute force)
  - OAuth2: 20 requests/minute (ป้องกัน OAuth abuse)
  - General API: 100 requests/minute (ป้องกัน excessive usage)
- **Automatic Token Refill**: Token regeneration ตาม time window
- **Memory Management**: Auto cleanup expired limiters
- **Thread-Safe**: Concurrent request handling

### Security Headers
- X-Content-Type-Options
- X-Frame-Options
- X-XSS-Protection
- Strict-Transport-Security
- Content-Security-Policy

### Password Security
- **bcrypt**: Password hashing
- **Cost Factor**: Configurable (default: 12)

## Database Schema

แอปใช้ SQLite database พร้อม GORM ORM:

### Core Tables
- `users` - ข้อมูลผู้ใช้ และ profile information
- `refresh_tokens` - Refresh tokens สำหรับ token renewal
- `oauth_states` - OAuth state management สำหรับ CSRF protection

### OAuth2 Tables
- `oauth2_clients` - OAuth2 client applications
- `oauth2_authorization_codes` - Authorization codes (short-lived)
- `oauth2_access_tokens` - Access tokens (longer-lived)

### Schema Details

#### Users Table
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255),
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  avatar VARCHAR(255),
  provider VARCHAR(50) DEFAULT 'local',
  provider_id VARCHAR(255),
  email_verified BOOLEAN DEFAULT false,
  is_active BOOLEAN DEFAULT true,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
```

#### OAuth2 Clients Table
```sql
CREATE TABLE oauth2_clients (
  id INTEGER PRIMARY KEY,
  client_id VARCHAR(255) UNIQUE NOT NULL,
  client_secret VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  redirect_uris JSON,
  scopes JSON,
  grant_types JSON,
  is_active BOOLEAN DEFAULT true,
  created_at DATETIME,
  updated_at DATETIME
);
```

## Monitoring & Analytics

### Rate Limit Monitoring
- Real-time rate limit status
- Per-IP token consumption tracking
- Active limiters monitoring
- Memory usage optimization

### API Information
- Complete endpoint documentation
- Feature list และ capabilities
- Version information
- Security features overview

## Development

### Project Structure
```
ex4-oauth2/
├── main.go                     # Application entry point
├── internal/
│   ├── auth/                   # JWT & OAuth utilities
│   │   └── jwt.go              # JWT, ID Token, JWK management
│   ├── config/                 # Configuration management
│   │   └── config.go
│   ├── database/               # Database layer
│   │   ├── repository.go       # User & token repositories
│   │   └── oauth2_repository.go # OAuth2 specific repositories
│   ├── handlers/               # HTTP handlers
│   │   ├── auth.go             # Authentication endpoints
│   │   ├── oauth.go            # OAuth2 server endpoints
│   │   └── oauth_client.go     # Client management
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go             # Authentication & security
│   │   └── ratelimit.go        # Rate limiting middleware
│   └── models/                 # Data models
│       ├── user.go             # User & auth models
│       └── oauth2.go           # OAuth2 specific models
├── .env.example               # Environment variables template
├── go.mod                     # Go module file
├── go.sum                     # Go dependencies
└── README.md                  # This documentation
```

### Core Components

#### Authentication Layer (`internal/auth/`)
- **JWT Service**: RS256/HMAC256 token generation and validation
- **ID Token**: Complete OIDC claims implementation with profile data
- **JWK Support**: Public key distribution for token verification
- **PKCE Utilities**: Code challenge/verifier generation and validation

#### Security Layer (`internal/middleware/`)
- **Token Bucket Rate Limiting**: Advanced rate limiting with automatic token refill
- **Authentication Middleware**: JWT validation and user context
- **Security Headers**: CORS, XSS protection, content security policies
- **Request ID**: Request tracing and logging correlation

#### Database Layer (`internal/database/`)
- **Repository Pattern**: Clean data access layer with GORM
- **Auto Migration**: Automatic database schema updates
- **Transaction Support**: Atomic operations for data consistency

### Testing & Development
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test specific package
go test ./internal/auth/

# Run in development mode
ENV=development go run main.go

# Build for development
go build -o auth-server main.go
```

### Building for Production
```bash
# Production build
CGO_ENABLED=1 go build -ldflags="-s -w" -o auth-server main.go

# Cross compilation for Linux
GOOS=linux GOARCH=amd64 go build -o auth-server-linux main.go

# Docker build (if using Docker)
docker build -t oauth2-server .
```

## Production Deployment

### Environment Configuration
```env
# Required
ENV=production
JWT_SECRET=your-super-secret-jwt-key-at-least-32-characters
DATABASE_URL=postgres://user:pass@localhost/dbname

# Optional with defaults
PORT=8080
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
RATE_LIMIT=100  # requests per minute per IP
```

### Production Checklist
- [ ] **Security**: Strong JWT secret (32+ characters)
- [ ] **HTTPS**: SSL/TLS certificates configured
- [ ] **Database**: Production database (PostgreSQL/MySQL)
- [ ] **CORS**: Proper allowed origins configuration
- [ ] **Rate Limiting**: Enabled and configured appropriately
- [ ] **Logging**: Structured logging with log levels
- [ ] **Monitoring**: Health checks and metrics collection
- [ ] **Backup**: Database backup strategy implemented
- [ ] **Environment**: All sensitive data in environment variables

### Performance Optimization
- Use connection pooling for database
- Enable HTTP/2 and compression
- Configure proper caching headers
- Set up load balancing if needed
- Monitor rate limiting effectiveness
- Use CDN for static content

### Security Best Practices
1. **JWT Security**: Use RS256 for production, rotate keys regularly
2. **Database**: Use prepared statements, enable query logging
3. **Network**: Firewall rules, VPN access for admin endpoints
4. **Monitoring**: Log failed authentication attempts, unusual activity
5. **Updates**: Keep dependencies updated, security patches applied

## License

MIT License