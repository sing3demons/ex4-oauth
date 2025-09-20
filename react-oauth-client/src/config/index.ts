import type { OAuthConfig, AppConfig } from '../types';

/**
 * OAuth2 Configuration
 * ตั้งค่าสำหรับเชื่อมต่อกับ OAuth2 Authorization Server
 */
export const oauthConfig: OAuthConfig = {
  // OAuth2 Server endpoints (ตรงกับ backend ที่เราสร้าง)
  authorizationEndpoint: 'http://localhost:8080/api/auth/oauth/authorize',
  tokenEndpoint: 'http://localhost:8080/api/auth/oauth/token',
  userInfoEndpoint: 'http://localhost:8080/api/auth/oauth/userinfo',
  
  // Client configuration
  clientId: 'WdJTw9_Ak51FZWt7znskk-GFKejDfsiS',
  
  // Redirect URIs (ต้อง match กับที่ register ใน OAuth2 server)
  redirectUri: 'http://localhost:5173/callback',
  postLogoutRedirectUri: 'http://localhost:5173',
  
  // OAuth2 scopes
  scope: 'openid email profile',
};

/**
 * API Configuration
 */
export const apiConfig = {
  baseURL: 'http://localhost:8080/api',
  timeout: 10000,
  withCredentials: false,
};

/**
 * WebSocket Configuration
 */
export const webSocketConfig = {
  url: 'ws://localhost:8080/api/ws/connect',
  reconnectInterval: 5000,
  maxReconnectAttempts: 5,
  heartbeatInterval: 30000,
};

/**
 * App Configuration
 */
export const appConfig: AppConfig = {
  apiBaseUrl: apiConfig.baseURL,
  oauth: oauthConfig,
  theme: {
    mode: 'light',
    primaryColor: '#3b82f6',
  },
  features: {
    notifications: true,
    adminPanel: true,
    userManagement: true,
  },
};

/**
 * Storage Keys
 */
export const storageKeys = {
  accessToken: 'access_token',
  refreshToken: 'refresh_token',
  idToken: 'id_token',
  codeVerifier: 'code_verifier',
  state: 'state',
  nonce: 'nonce',
  user: 'user',
  theme: 'theme',
} as const;

/**
 * Routes Configuration
 */
export const routes = {
  home: '/',
  login: '/login',
  callback: '/callback',
  dashboard: '/dashboard',
  profile: '/profile',
  settings: '/settings',
  admin: '/admin',
  adminUsers: '/admin/users',
  adminClients: '/admin/clients',
  adminAudit: '/admin/audit',
  adminSecurity: '/admin/security',
  notifications: '/notifications',
  unauthorized: '/unauthorized',
  notFound: '/404',
} as const;

/**
 * API Endpoints
 */
export const endpoints = {
  // Authentication
  auth: {
    register: '/auth/register',
    login: '/auth/login',
    refresh: '/auth/refresh',
    logout: '/auth/logout',
    profile: '/auth/profile',
    changePassword: '/auth/change-password',
  },
  
  // OAuth2
  oauth: {
    authorize: '/auth/oauth/authorize',
    token: '/auth/oauth/token',
    userinfo: '/auth/oauth/userinfo',
    discovery: '/auth/oauth/.well-known/openid_configuration',
    jwks: '/auth/oauth/.well-known/jwks.json',
  },
  
  // Users
  users: {
    list: '/users',
    profile: '/auth/profile',
  },
  
  // Admin
  admin: {
    stats: '/admin/stats',
    users: '/admin/users',
    clients: '/admin/clients',
    audit: '/admin/audit',
    security: '/admin/audit/security-alerts',
  },
  
  // Notifications
  notifications: {
    list: '/notifications',
    stats: '/notifications/stats',
    markRead: '/notifications/:id/read',
    markAllRead: '/notifications/read-all',
    delete: '/notifications/:id',
  },
  
  // WebSocket
  websocket: {
    connect: '/ws/connect',
    public: '/ws/public',
  },
} as const;

/**
 * Environment Variables
 */
export const env = {
  NODE_ENV: import.meta.env.NODE_ENV || 'development',
  VITE_API_BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  VITE_OAUTH_CLIENT_ID: import.meta.env.VITE_OAUTH_CLIENT_ID || 'WdJTw9_Ak51FZWt7znskk-GFKejDfsiS',
  VITE_OAUTH_REDIRECT_URI: import.meta.env.VITE_OAUTH_REDIRECT_URI || 'http://localhost:5173/callback',
  VITE_WS_URL: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/ws/connect',
} as const;

/**
 * Development mode check
 */
export const isDevelopment = env.NODE_ENV === 'development';
export const isProduction = env.NODE_ENV === 'production';

/**
 * Token expiration warning time (minutes)
 */
export const TOKEN_EXPIRY_WARNING_TIME = 5;

/**
 * Auto refresh token before expiry (minutes)
 */
export const AUTO_REFRESH_BEFORE_EXPIRY = 10;

/**
 * Maximum retry attempts for API calls
 */
export const MAX_RETRY_ATTEMPTS = 3;

/**
 * Default pagination settings
 */
export const pagination = {
  defaultPage: 1,
  defaultLimit: 20,
  maxLimit: 100,
} as const;

/**
 * Notification settings
 */
export const notificationSettings = {
  defaultDuration: 5000,
  errorDuration: 8000,
  successDuration: 3000,
  position: 'top-right',
} as const;

/**
 * Theme settings
 */
export const themeSettings = {
  defaultTheme: 'light',
  storageKey: 'theme-preference',
} as const;

/**
 * Security settings
 */
export const securitySettings = {
  maxLoginAttempts: 5,
  lockoutDuration: 900, // 15 minutes in seconds
  sessionTimeout: 3600, // 1 hour in seconds
  passwordMinLength: 8,
} as const;