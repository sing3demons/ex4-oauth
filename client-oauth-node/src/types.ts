// OAuth2 Types
export interface OAuth2Config {
  clientId: string;
  clientSecret: string;
  serverUrl: string;
  redirectUri: string;
  scope?: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  token_type: string;
  expires_in: number;
  scope?: string;
}

export interface UserInfo {
  sub: string;
  email?: string;
  email_verified?: boolean;
  name?: string;
  given_name?: string;
  family_name?: string;
  picture?: string;
  preferred_username?: string;
}

export interface OIDCDiscovery {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri?: string;
  scopes_supported: string[];
  response_types_supported: string[];
  grant_types_supported: string[];
  code_challenge_methods_supported?: string[];
  claims_supported?: string[];
}

// Error Types
export interface OAuth2Error {
  error: string;
  error_description?: string;
  error_uri?: string;
}

export interface AppConfig {
  port: number;
  nodeEnv: string;
  oauth: OAuth2Config;
  app: {
    name: string;
    url: string;
  };
}