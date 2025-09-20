/**
 * OAuth2 Configuration Interface
 */
export interface OAuthConfig {
  authorizationEndpoint: string;
  tokenEndpoint: string;
  userInfoEndpoint: string;
  clientId: string;
  redirectUri: string;
  scope: string;
  postLogoutRedirectUri: string;
}

/**
 * Login Credentials
 */
export interface LoginCredentials {
  email: string;
  password: string;
}

/**
 * Registration Data
 */
export interface RegisterData {
  email: string;
  password: string;
  username: string;
  first_name: string;
  last_name: string;
}

/**
 * Password Change Data
 */
export interface ChangePasswordData {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

/**
 * Password Reset Data
 */
export interface PasswordResetData {
  token: string;
  new_password: string;
  confirm_password: string;
}

/**
 * OAuth2 Token Response (RFC 6749)
 */
export interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  token_type: string;
  expires_in: number;
  scope: string;
}

/**
 * OAuth2 Error Response (RFC 6749)
 */
export interface OAuthError {
  error: string;
  error_description?: string;
  error_uri?: string;
  state?: string;
}

/**
 * User Profile from UserInfo endpoint (OIDC)
 */
export interface UserProfile {
  sub: string;                    // Subject - unique user identifier
  email: string;                  // Email address
  email_verified?: boolean;       // Email verification status
  name?: string;                  // Full name
  given_name?: string;           // First name
  family_name?: string;          // Last name
  preferred_username?: string;    // Username
  picture?: string;              // Profile picture URL
  role?: string;                 // User role (custom claim)
  iat?: number;                  // Issued at
  exp?: number;                  // Expires at
  aud?: string | string[];       // Audience
  iss?: string;                  // Issuer
}

/**
 * Authentication State
 */
export interface AuthState {
  user: UserProfile | null;
  loading: boolean;
  error: string | null;
  isAuthenticated: boolean;
}

/**
 * OAuth2 PKCE Parameters
 */
export interface PKCEParams {
  codeVerifier: string;
  codeChallenge: string;
  state: string;
  nonce: string;
}

/**
 * Authorization Request Parameters
 */
export interface AuthorizationParams {
  response_type: string;
  client_id: string;
  redirect_uri: string;
  scope: string;
  state: string;
  code_challenge: string;
  code_challenge_method: string;
  nonce: string;
}

/**
 * Token Request Parameters
 */
export interface TokenRequestParams {
  grant_type: string;
  code: string;
  redirect_uri: string;
  client_id: string;
  code_verifier: string;
}

/**
 * Refresh Token Request Parameters
 */
export interface RefreshTokenParams {
  grant_type: string;
  refresh_token: string;
  client_id: string;
  scope?: string;
}

/**
 * API Response wrapper
 */
export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
}

/**
 * API Error Response
 */
export interface ApiError {
  error: string;
  error_description?: string;
  message?: string;
  status?: number;
}

/**
 * WebSocket Message Types
 */
export type NotificationType = 
  | 'SECURITY_ALERT'
  | 'SYSTEM_NOTIFICATION'
  | 'USER_NOTIFICATION'
  | 'TOKEN_EXPIRY'
  | 'ADMIN_ALERT'
  | 'AUDIT_ALERT'
  | 'COMPLIANCE_ALERT';

export type NotificationSeverity = 'low' | 'medium' | 'high';

/**
 * WebSocket Notification Message
 */
export interface WebSocketMessage {
  id: string;
  type: NotificationType;
  title: string;
  message: string;
  severity: NotificationSeverity;
  channel: string;
  user_id?: string;
  metadata?: Record<string, any>;
  created_at: string;
  expires_at?: string;
}

/**
 * Notification State
 */
export interface NotificationState {
  notifications: WebSocketMessage[];
  unreadCount: number;
  connected: boolean;
  error: string | null;
}

/**
 * Admin Dashboard Stats
 */
export interface DashboardStats {
  total_users: number;
  active_users: number;
  verified_users: number;
  total_clients: number;
  active_clients: number;
  total_tokens: number;
  expired_tokens: number;
  pending_verifications: number;
  recent_logins_24h: number;
}

/**
 * User Management
 */
export interface User {
  id: number;
  email: string;
  username: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  email_verified: boolean;
  is_active: boolean;
  role: 'user' | 'moderator' | 'admin';
  created_at: string;
  updated_at: string;
}

/**
 * OAuth2 Client
 */
export interface OAuth2Client {
  id: number;
  client_id: string;
  name: string;
  description?: string;
  redirect_uris: string[];
  scopes: string[];
  grant_types: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/**
 * Audit Log Entry
 */
export interface AuditLog {
  id: number;
  user_id?: number;
  action: string;
  resource_type: string;
  resource_id?: string;
  ip_address?: string;
  user_agent?: string;
  success: boolean;
  error_message?: string;
  category: string;
  risk_level: 'low' | 'medium' | 'high';
  created_at: string;
}

/**
 * Security Alert
 */
export interface SecurityAlert {
  id: number;
  alert_type: string;
  severity: 'low' | 'medium' | 'high';
  status: 'open' | 'resolved' | 'dismissed';
  user_id?: number;
  ip_address?: string;
  threat_level: number;
  description: string;
  evidence?: string;
  created_at: string;
  resolved_at?: string;
}

/**
 * Form validation errors
 */
export interface ValidationErrors {
  [field: string]: string[];
}

/**
 * HTTP Method types
 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

/**
 * Route configuration
 */
export interface RouteConfig {
  path: string;
  component: React.ComponentType;
  protected?: boolean;
  adminOnly?: boolean;
  title?: string;
}

/**
 * Theme configuration
 */
export interface ThemeConfig {
  mode: 'light' | 'dark';
  primaryColor: string;
}

/**
 * App configuration
 */
export interface AppConfig {
  apiBaseUrl: string;
  oauth: OAuthConfig;
  theme: ThemeConfig;
  features: {
    notifications: boolean;
    adminPanel: boolean;
    userManagement: boolean;
  };
}