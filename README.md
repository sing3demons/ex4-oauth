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
- � Email Verification System
- 👨‍💼 Admin Dashboard & User Management
- 🔍 Comprehensive Audit Logging
- 🚨 Security Event Monitoring
- �📊 Compliance Reporting
- 🎯 Role-Based Access Control (RBAC)
- 📈 Monitoring & Analytics

## คุณสมบัติหลัก

### Authentication & Authorization
- สมัครสมาชิกและเข้าสู่ระบบด้วย email/password
- Email verification system พร้อม SMTP configuration
- Password reset ผ่าน email
- JWT access tokens และ refresh tokens (RS256/HMAC256)
- OAuth2 Authorization Code Grant พร้อม PKCE
- OpenID Connect สำหรับ SSO พร้อม ID Token
- Consent screen สำหรับการขออนุญาต
- Rate limiting ป้องกัน brute force และ API abuse
- Role-Based Access Control (user, moderator, admin)

### Admin Dashboard & Management
- Complete admin dashboard พร้อม statistics
- User management (CRUD operations)
- Role assignment และ status management
- Activity logging และ monitoring
- System events tracking
- User activity audit trail
- Admin-only endpoints พร้อม middleware protection

### Audit & Compliance System
- Comprehensive audit logging สำหรับทุก API requests
- Security event detection และ alerting
- Failed authentication tracking
- Suspicious activity monitoring (SQL injection, bot detection)
- Compliance reporting (GDPR, HIPAA, SOX)
- Risk assessment และ threat level analysis
- Automatic audit trail generation
- Log retention และ cleanup policies

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
- 📧 Email verification พร้อม customizable templates
- 🔐 RBAC system with admin/moderator/user roles
- 🕵️ SQL injection และ suspicious activity detection
- 🚨 Real-time security alerting
- 📋 Comprehensive audit logging
- 🛡️ Security headers (XSS, CSRF, Content Security Policy)

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

# Email Configuration (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourapp.com
FROM_NAME=YourApp

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
- `POST /api/auth/logout-all` - ออกจากระบบทุก device
- `GET /api/auth/profile` - ดู profile
- `PUT /api/auth/profile` - แก้ไข profile
- `POST /api/auth/change-password` - เปลี่ยนรหัสผ่าน

### Email Verification
- `POST /api/auth/send-verification` - ส่ง email verification
- `GET /api/auth/verify-email` - verify email (via link)
- `POST /api/auth/verify-email` - verify email (via form)
- `POST /api/auth/password-reset` - ส่ง password reset email
- `POST /api/auth/reset-password` - reset password ผ่าน token

### Email Management (Protected)
- `GET /api/email/stats` - ดู email statistics
- `POST /api/email/cleanup` - ลบ expired email tokens

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

### OAuth2 Token Introspection & Revocation
- `POST /api/oauth2/introspect` - Token introspection (RFC 7662)
- `POST /api/oauth2/revoke` - Token revocation (RFC 7009)
- `GET /api/oauth2/tokeninfo` - Token information (debugging)

### User Management
- `GET /api/users` - ดูรายการผู้ใช้

### Admin Dashboard (Admin Only)
- `GET /api/admin/stats` - Dashboard statistics
- `GET /api/admin/users` - จัดการผู้ใช้ (ดู, filter, search)
- `GET /api/admin/users/:id/activity` - ดู user activity logs
- `GET /api/admin/events` - ดู system events
- `PUT /api/admin/users/:id/status` - เปลี่ยน user status (active/suspended/banned)
- `PUT /api/admin/users/:id/role` - เปลี่ยน user role (user/moderator/admin)
- `DELETE /api/admin/users/:id` - ลบ user
- `POST /api/admin/cleanup-logs` - ลบ logs เก่า
- `GET /api/admin/system-info` - ดู system information

### Token Management (Admin Only)
- `GET /api/admin/tokens/stats/:userID` - ดู token usage statistics

### Audit & Compliance (Admin Only)
- `GET /api/admin/audit/logs` - ดู audit logs พร้อม filtering
- `GET /api/admin/audit/search` - ค้นหา audit logs
- `GET /api/admin/audit/users/:id/trail` - ดู audit trail ของ user
- `GET /api/admin/audit/resources/:type/:id/trail` - ดู audit trail ของ resource
- `GET /api/admin/audit/high-risk` - ดู high-risk activities
- `GET /api/admin/audit/security-alerts` - ดู security alerts
- `POST /api/admin/audit/security-alerts/:id/resolve` - แก้ไข security alert
- `POST /api/admin/audit/compliance/reports` - สร้าง compliance report
- `POST /api/admin/audit/cleanup` - ลบ audit logs เก่า

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

### 5. Email Verification

```bash
# ส่ง verification email
curl -X POST http://localhost:8080/api/auth/send-verification \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Verify email ผ่าน token
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "token": "VERIFICATION_TOKEN"
  }'
```

### 6. Password Reset

```bash
# ส่ง password reset email
curl -X POST http://localhost:8080/api/auth/password-reset \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'

# Reset password ผ่าน token
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "RESET_TOKEN",
    "new_password": "newpassword123"
  }'
```

### 7. Token Introspection & Revocation

#### Token Introspection (RFC 7662)
```bash
# Introspect access token
curl -X POST http://localhost:8080/api/oauth2/introspect \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d 'token=YOUR_ACCESS_TOKEN&token_type_hint=access_token'
```

**Response สำหรับ active token:**
```json
{
  "active": true,
  "scope": "openid email profile",
  "client_id": "your-client-id",
  "username": "john_doe",
  "token_type": "Bearer",
  "exp": 1731145200,
  "iat": 1731144300,
  "sub": "123",
  "aud": ["your-client-id"],
  "iss": "ex4-oauth2",
  "user_id": 123,
  "email": "john@example.com",
  "role": "user",
  "provider": "local"
}
```

**Response สำหรับ inactive token:**
```json
{
  "active": false
}
```

#### Token Revocation (RFC 7009)
```bash
# Revoke refresh token
curl -X POST http://localhost:8080/api/oauth2/revoke \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d 'token=YOUR_REFRESH_TOKEN'
```

#### Token Information (Debug Endpoint)
```bash
# Get token info using Bearer authentication
curl -X GET http://localhost:8080/api/oauth2/tokeninfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 8. Admin Dashboard Features

#### Get Dashboard Statistics (Admin)
```bash
curl -X GET http://localhost:8080/api/admin/stats \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

**Response:**
```json
{
  "total_users": 1250,
  "active_users": 1180,
  "verified_users": 980,
  "total_clients": 15,
  "active_clients": 12,
  "total_tokens": 3450,
  "expired_tokens": 450,
  "pending_verifications": 23,
  "recent_logins_24h": 156
}
```

#### Token Usage Statistics (Admin)
```bash
# ดู token usage statistics ของ user
curl -X GET http://localhost:8080/api/admin/tokens/stats/123?days=30 \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

#### User Management (Admin)
```bash
# ดูรายการ users พร้อม filtering
curl -X GET "http://localhost:8080/api/admin/users?limit=20&offset=0&role=user&status=active" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"

# เปลี่ยน user role
curl -X PUT http://localhost:8080/api/admin/users/123/role \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -d '{
    "role": "moderator"
  }'

# เปลี่ยน user status
curl -X PUT http://localhost:8080/api/admin/users/123/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -d '{
    "is_active": false,
    "reason": "Policy violation"
  }'
```

### 8. Audit & Compliance Features

#### Get Audit Logs
```bash
curl -X GET "http://localhost:8080/api/admin/audit/logs?limit=50&category=authentication&risk_level=medium" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 123,
      "action": "login",
      "resource_type": "authentication",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "category": "authentication",
      "risk_level": "low",
      "created_at": "2025-09-20T20:30:00Z"
    }
  ],
  "pagination": {
    "total": 1250,
    "limit": 50,
    "offset": 0
  }
}
```

#### Search Audit Logs
```bash
curl -X GET "http://localhost:8080/api/admin/audit/search?query=failed+login&limit=20" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

#### Get Security Alerts
```bash
curl -X GET "http://localhost:8080/api/admin/audit/security-alerts?status=open&severity=high" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "alert_type": "suspicious_activity",
      "severity": "high",
      "status": "open",
      "threat_level": 8,
      "description": "Multiple failed login attempts detected",
      "ip_address": "192.168.1.100",
      "created_at": "2025-09-20T20:30:00Z"
    }
  ]
}
```

#### Generate Compliance Report
```bash
curl -X POST http://localhost:8080/api/admin/audit/compliance/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -d '{
    "report_type": "gdpr",
    "period": "last_month"
  }'
```

### 9. ID Token (OIDC)

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

### 10. JWK (JSON Web Key)

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

### 11. Rate Limiting

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

### 12. Monitoring Rate Limits

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

### Email System Tables
- `email_verifications` - Email verification tokens และ password reset
- `email_templates` - Customizable email templates

### Admin & Audit Tables
- `user_activities` - User activity logging
- `system_events` - System-level event logging
- `audit_logs` - Comprehensive audit trail สำหรับทุก actions
- `security_alerts` - Security event alerts และ threat detection
- `compliance_reports` - Compliance และ audit reports

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
  role VARCHAR(50) DEFAULT 'user',
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

#### Audit Logs Table
```sql
CREATE TABLE audit_logs (
  id INTEGER PRIMARY KEY,
  user_id INTEGER,
  session_id VARCHAR(255),
  action VARCHAR(255) NOT NULL,
  resource_type VARCHAR(255) NOT NULL,
  resource_id VARCHAR(255),
  old_values TEXT,
  new_values TEXT,
  ip_address VARCHAR(45),
  user_agent TEXT,
  request_method VARCHAR(10),
  request_path VARCHAR(255),
  request_headers TEXT,
  response_status INTEGER,
  success BOOLEAN DEFAULT true,
  error_message TEXT,
  duration BIGINT,
  severity VARCHAR(50) DEFAULT 'info',
  category VARCHAR(50) NOT NULL,
  risk_level VARCHAR(50) DEFAULT 'low',
  compliance_flags VARCHAR(255),
  geo_location VARCHAR(255),
  device_info VARCHAR(255),
  created_at DATETIME
);
```

#### Security Alerts Table
```sql
CREATE TABLE security_alerts (
  id INTEGER PRIMARY KEY,
  alert_type VARCHAR(255) NOT NULL,
  severity VARCHAR(50) NOT NULL,
  status VARCHAR(50) DEFAULT 'open',
  user_id INTEGER,
  ip_address VARCHAR(45),
  user_agent TEXT,
  threat_level INTEGER DEFAULT 1,
  description TEXT NOT NULL,
  evidence TEXT,
  metadata TEXT,
  triggered_by VARCHAR(255),
  resolved_by INTEGER,
  resolved_at DATETIME,
  resolution TEXT,
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

### Admin Dashboard Analytics
- User statistics และ demographics
- Authentication success/failure rates
- OAuth2 client usage patterns
- System resource utilization

### Audit & Security Monitoring
- Real-time security threat detection
- Failed authentication pattern analysis
- Suspicious activity alerting
- Compliance violation tracking
- Geographic access pattern monitoring

### API Information
- Complete endpoint documentation
- Feature list และ capabilities
- Version information
- Security features overview

## Real-time Notifications

The OAuth2 server provides a comprehensive real-time notification system using WebSocket connections.

### WebSocket Connection

Connect to real-time notifications:

```bash
# Authenticated WebSocket connection
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     --header "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     http://localhost:8080/api/ws/connect

# Anonymous WebSocket connection (limited features)
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://localhost:8080/api/ws/public
```

### Notification Management

Get user notifications:

```bash
curl -X GET "http://localhost:8080/api/notifications" \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Mark notification as read:

```bash
curl -X PUT "http://localhost:8080/api/notifications/123/read" \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Mark all notifications as read:

```bash
curl -X PUT "http://localhost:8080/api/notifications/read-all" \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Get notification statistics:

```bash
curl -X GET "http://localhost:8080/api/notifications/stats" \
     -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Admin Notifications

Get admin notifications:

```bash
curl -X GET "http://localhost:8080/api/admin/notifications" \
     -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Get admin notification statistics:

```bash
curl -X GET "http://localhost:8080/api/admin/notifications/stats" \
     -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Test notification system:

```bash
curl -X POST "http://localhost:8080/api/admin/notifications/test" \
     -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "type": "SECURITY_ALERT",
       "message": "Test security alert",
       "severity": "high"
     }'
```

### WebSocket Message Format

The WebSocket connection sends messages in the following format:

```json
{
  "id": "notification_id",
  "type": "SECURITY_ALERT",
  "title": "Security Alert",
  "message": "Multiple failed login attempts detected",
  "severity": "high",
  "channel": "security",
  "user_id": "user123",
  "metadata": {
    "attempts": 5,
    "ip": "192.168.1.100"
  },
  "created_at": "2023-12-01T10:00:00Z",
  "expires_at": "2023-12-01T11:00:00Z"
}
```

### Channel Subscriptions

Subscribe to specific notification channels:

```javascript
// JavaScript WebSocket client example
const ws = new WebSocket('ws://localhost:8080/api/ws/connect', [], {
  headers: {
    'Authorization': 'Bearer ' + accessToken
  }
});

ws.onopen = function() {
  // Subscribe to channels
  ws.send(JSON.stringify({
    action: 'subscribe',
    channels: ['security', 'system', 'user']
  }));
};

ws.onmessage = function(event) {
  const notification = JSON.parse(event.data);
  console.log('Received notification:', notification);
};
```

### Notification Types

The system supports various notification types:

- **SECURITY_ALERT**: Security-related alerts (failed logins, suspicious activity)
- **SYSTEM_NOTIFICATION**: System-wide announcements
- **USER_NOTIFICATION**: User-specific notifications
- **TOKEN_EXPIRY**: Token expiration warnings
- **ADMIN_ALERT**: Administrative alerts
- **AUDIT_ALERT**: Audit-related notifications
- **COMPLIANCE_ALERT**: Compliance-related notifications

### Security Features

- **Authentication**: WebSocket connections require valid JWT tokens
- **Channel-based Access**: Users only receive notifications for authorized channels
- **Rate Limiting**: Prevents notification spam
- **Message Encryption**: All WebSocket communications are secure
- **Audit Integration**: All notifications are logged for audit purposes

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
│   │   ├── oauth_client.go     # Client management
│   │   ├── email.go            # Email verification handlers
│   │   ├── admin.go            # Admin dashboard handlers
│   │   └── audit.go            # Audit & compliance handlers
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go             # Authentication & security
│   │   ├── ratelimit.go        # Rate limiting middleware
│   │   ├── admin.go            # Admin access control
│   │   └── audit.go            # Audit logging middleware
│   ├── models/                 # Data models
│   │   ├── user.go             # User & auth models
│   │   └── oauth2.go           # OAuth2 specific models
│   └── services/               # Business logic layer
│       ├── auth.go             # Authentication services
│       ├── email.go            # Email services
│       ├── admin.go            # Admin services
│       └── audit.go            # Audit services
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

#### Business Logic Layer (`internal/services/`)
- **Auth Service**: User authentication and session management
- **Email Service**: SMTP integration and template management
- **Admin Service**: User management and administrative operations
- **Audit Service**: Comprehensive logging and compliance reporting

#### Security Layer (`internal/middleware/`)
- **Token Bucket Rate Limiting**: Advanced rate limiting with automatic token refill
- **Authentication Middleware**: JWT validation and user context
- **Admin Middleware**: Role-based access control (RBAC)
- **Audit Middleware**: Automatic request logging and security monitoring
- **Security Headers**: CORS, XSS protection, content security policies
- **Request ID**: Request tracing and logging correlation

#### Database Layer (`internal/database/`)
- **Repository Pattern**: Clean data access layer with GORM
- **Auto Migration**: Automatic database schema updates
- **Transaction Support**: Atomic operations for data consistency
- **Audit Repository**: Specialized audit and compliance data operations

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
- [ ] **Email**: SMTP configuration for production
- [ ] **Admin Access**: Secure admin account creation
- [ ] **Audit Retention**: Log retention policies configured
- [ ] **Compliance**: GDPR/data protection compliance setup
- [ ] **Logging**: Structured logging with log levels
- [ ] **Monitoring**: Health checks and metrics collection
- [ ] **Backup**: Database backup strategy implemented
- [ ] **Environment**: All sensitive data in environment variables
- [ ] **Security Headers**: HTTPS, CSP, HSTS properly configured

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
4. **Email Security**: Use app passwords, verify SMTP TLS
5. **Admin Access**: Strong passwords, 2FA when possible
6. **Audit Logs**: Regular log review, automated threat detection
7. **Data Protection**: GDPR compliance, data retention policies
8. **Monitoring**: Log failed authentication attempts, unusual activity
9. **Updates**: Keep dependencies updated, security patches applied
10. **Backup**: Regular automated backups with encryption

## License

MIT License