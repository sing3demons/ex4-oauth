# NodeJS OAuth2 Client

Node.js TypeScript application ที่ทำงานเป็น OAuth2 Client กับ Go OAuth2 Authorization Server

## 🚀 Features

- **OAuth2 Authorization Code Flow** with PKCE
- **OpenID Connect (OIDC)** integration
- **Session Management** with Express sessions
- **Token Management** (access/refresh tokens)
- **Protected Routes** with middleware
- **TypeScript** strong typing
- **Auto Token Refresh** middleware

## 📋 Prerequisites

- Node.js 18+ 
- Go OAuth2 Server running on `http://localhost:8080`
- MongoDB running (for OAuth2 server)

## 🛠️ Installation

```bash
# Navigate to client directory
cd nodejs-oauth-client

# Install dependencies
npm install

# Copy environment variables
cp .env.example .env

# Edit .env with your configuration
nano .env
```

## ⚙️ Configuration

Edit `.env` file:

```env
PORT=3001
NODE_ENV=development

# OAuth2 Server Configuration
OAUTH_SERVER_URL=http://localhost:8080
OAUTH_CLIENT_ID=4xsmjvVypydAOUFZrDxwn12J-rWdFIYa
OAUTH_CLIENT_SECRET=9mYxPS4p_G-LoxaYnT-szafwVoRwh_BLe9zh1TGcGBZgi0OJISppd4U12KK_2tsh
OAUTH_REDIRECT_URI=http://localhost:3001/auth/callback

# Session Configuration
SESSION_SECRET=your-super-secret-session-key-change-in-production

# Application Configuration
APP_NAME=NodeJS OAuth Client
APP_URL=http://localhost:3001
```

## 🚀 Running

### Development
```bash
npm run dev
```

### Production Build
```bash
npm run build
npm start
```

## 📍 API Endpoints

### Public Endpoints
- `GET /` - Home page with auth status
- `GET /health` - Health check
- `GET /api/info` - Application info

### Authentication Endpoints
- `GET /auth/login` - Start OAuth2 login flow
- `GET /auth/callback` - OAuth2 callback handler
- `POST /auth/logout` - Logout and clear session

### Protected Endpoints (require authentication)
- `GET /api/profile` - Get user profile
- `GET /api/protected` - Example protected endpoint

## 🔄 OAuth2 Flow

1. **Start Login**: Visit `/auth/login`
2. **Authorization**: Redirected to OAuth2 server
3. **User Login**: Login on OAuth2 server
4. **Callback**: Return to `/auth/callback` with authorization code
5. **Token Exchange**: Exchange code for access token
6. **User Info**: Fetch user information
7. **Session**: Store user data in session

## 🔐 Security Features

- **PKCE (Proof Key for Code Exchange)** for authorization code flow
- **State Parameter** to prevent CSRF attacks
- **Nonce Parameter** for OpenID Connect
- **Session Management** with secure cookies
- **Auto Token Refresh** before expiration
- **Token Revocation** on logout

## 📖 Usage Examples

### 1. Basic Authentication Flow

```bash
# Start login
curl http://localhost:3001/auth/login

# After login, check profile
curl -b cookies.txt http://localhost:3001/api/profile

# Access protected endpoint
curl -b cookies.txt http://localhost:3001/api/protected
```

### 2. Using with Frontend

```javascript
// Login
window.location.href = 'http://localhost:3001/auth/login';

// Check authentication status
fetch('http://localhost:3001/', { credentials: 'include' })
  .then(res => res.json())
  .then(data => console.log('Auth status:', data.authenticated));

// Access protected API
fetch('http://localhost:3001/api/profile', { credentials: 'include' })
  .then(res => res.json())
  .then(user => console.log('User:', user));
```

## 🏗️ Architecture

```
┌─────────────────┐    OAuth2 Flow    ┌─────────────────┐
│   Browser/App   │ ◄────────────────► │   NodeJS Client │
└─────────────────┘                    └─────────────────┘
                                              │
                                              │ OAuth2/OIDC
                                              ▼
                                    ┌─────────────────┐
                                    │   Go OAuth2     │
                                    │     Server      │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │    MongoDB      │
                                    └─────────────────┘
```

## 📁 Project Structure

```
nodejs-oauth-client/
├── src/
│   ├── index.ts        # Main application
│   ├── config.ts       # Configuration management
│   ├── types.ts        # TypeScript type definitions
│   ├── oauth.ts        # OAuth2 service implementation
│   ├── middleware.ts   # Express middleware
│   └── routes.ts       # API routes
├── package.json
├── tsconfig.json
├── .env
└── README.md
```

## 🔧 Development

### Type Checking
```bash
npm run type-check
```

### Building
```bash
npm run build
```

### Cleaning
```bash
npm run clean
```

## ⚠️ Important Notes

1. **Client Credentials**: Store securely in production
2. **Session Secret**: Use strong random secret in production
3. **HTTPS**: Use HTTPS in production for secure cookies
4. **Error Handling**: Implement proper error logging
5. **Token Storage**: Consider secure token storage alternatives

## 🐛 Troubleshooting

### Common Issues

1. **"OAuth2 Error: invalid_client"**
   - Check `OAUTH_CLIENT_ID` and `OAUTH_CLIENT_SECRET`
   - Verify client is registered in OAuth2 server

2. **"Invalid redirect_uri"**
   - Ensure `OAUTH_REDIRECT_URI` matches registered URI
   - Check OAuth2 server client configuration

3. **"Session not found"**
   - Check session configuration
   - Verify cookies are being sent

4. **"Token expired"**
   - Implement auto-refresh middleware
   - Check token expiration handling

## 🎯 Next Steps

- [ ] Add JWT validation
- [ ] Implement logout from OAuth2 server
- [ ] Add API rate limiting
- [ ] Implement database session store
- [ ] Add frontend UI
- [ ] Add comprehensive error handling
- [ ] Add monitoring and logging

## 📄 License

MIT License