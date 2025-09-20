# OAuth2 Authorization Server

แอปพลิเคชัน backend OAuth2 Authorization Server ที่สมบูรณ์ พร้อมการรองรับ:
- 📝 User Registration และ Login
- 🔐 JWT Authentication  
- 🛡️ OAuth2 Authorization Server
- 🆔 OpenID Connect (OIDC)
- 🔒 PKCE (Proof Key for Code Exchange)
- 👤 Profile Management
- ♻️ Token Refresh
- ⚙️ OAuth2 Client Management

## คุณสมบัติหลัก

### Authentication & Authorization
- สมัครสมาชิกและเข้าสู่ระบบด้วย email/password
- JWT access tokens และ refresh tokens
- OAuth2 Authorization Code Grant พร้อม PKCE
- OpenID Connect สำหรับ SSO
- Consent screen สำหรับการขออนุญาต

### OAuth2 Server Endpoints
- `/api/auth/oauth/authorize` - Authorization endpoint
- `/api/auth/oauth/token` - Token endpoint  
- `/api/auth/oauth/userinfo` - UserInfo endpoint
- `/api/auth/oauth/.well-known/openid_configuration` - Discovery endpoint

### Client Management
- สร้าง OAuth2 clients
- จัดการ redirect URIs และ scopes
- Client credentials management

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

### Authentication
- `POST /api/auth/register` - สมัครสมาชิก
- `POST /api/auth/login` - เข้าสู่ระบบ
- `POST /api/auth/refresh` - refresh token
- `POST /api/auth/logout` - ออกจากระบบ
- `GET /api/auth/profile` - ดู profile
- `PUT /api/auth/profile` - แก้ไข profile
- `POST /api/auth/change-password` - เปลี่ยนรหัสผ่าน

### OAuth2 Authorization Server
- `GET /api/auth/oauth/authorize` - Authorization endpoint
- `POST /api/auth/oauth/token` - Token endpoint
- `POST /api/auth/oauth/consent` - Consent endpoint
- `GET /api/auth/oauth/userinfo` - UserInfo endpoint
- `GET /api/auth/oauth/.well-known/openid_configuration` - OIDC Discovery

### OAuth2 Client Management
- `POST /api/oauth2/clients` - สร้าง OAuth2 client
- `GET /api/oauth2/clients` - ดูรายการ clients
- `GET /api/oauth2/clients/:id` - ดู client ตาม ID
- `PUT /api/oauth2/clients/:id` - แก้ไข client
- `DELETE /api/oauth2/clients/:id` - ลบ client

### User Management
- `GET /api/users` - ดูรายการผู้ใช้

### System
- `GET /health` - Health check
- `GET /api/info` - API information

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

#### ขั้นตอนที่ 3: UserInfo Request

```bash
curl -X GET http://localhost:8080/api/auth/oauth/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

## PKCE Implementation

แอปนี้รองรับ PKCE (Proof Key for Code Exchange) สำหรับความปลอดภัยเพิ่มเติม:

1. สร้าง `code_verifier` (43-128 characters)
2. สร้าง `code_challenge` จาก SHA256(code_verifier) และ base64url encode
3. ส่ง `code_challenge` และ `code_challenge_method=S256` ใน authorization request
4. ส่ง `code_verifier` ใน token request

## OpenID Connect (OIDC)

รองรับ OIDC discovery endpoint:

```bash
curl http://localhost:8080/api/auth/oauth/.well-known/openid_configuration
```

## Database Schema

แอปใช้ SQLite database พร้อม GORM ORM:

- `users` - ข้อมูลผู้ใช้
- `oauth2_clients` - OAuth2 client applications
- `oauth2_authorization_codes` - Authorization codes
- `oauth2_access_tokens` - Access tokens
- `refresh_tokens` - Refresh tokens
- `oauth_states` - OAuth state management

## Security Features

- 🔐 Password hashing ด้วย bcrypt
- 🛡️ CORS protection
- 🔒 Security headers
- ⚡ Rate limiting (พร้อมใช้งาน)
- 🔑 PKCE support
- 🆔 OIDC compliance
- 🎯 Secure token storage

## Development

### Directory Structure

```
ex4-oauth2/
├── internal/
│   ├── auth/           # JWT และ OAuth2 utilities
│   ├── config/         # Configuration management
│   ├── database/       # Database repositories
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
├── main.go            # Entry point
├── go.mod
└── README.md
```

### Testing

API ทดสอบได้ผ่าน:
- Postman
- curl commands
- HTTP clients
- OAuth2 testing tools

## Production Deployment

สำหรับ production:

1. ตั้งค่า `ENV=production`
2. ใช้ database จริง (PostgreSQL/MySQL)
3. ตั้งค่า proper CORS origins
4. ใช้ strong JWT secrets
5. เปิดใช้ rate limiting
6. ใช้ HTTPS
7. ตั้งค่า proper logging

## License

MIT License