import { PKCEUtils, TokenUtils, StorageUtils, URLUtils } from '../utils/oauth';
import { apiClient } from './apiClient';
import { oauthConfig } from '../config';
import type { TokenResponse, User, LoginCredentials, RegisterData } from '../types';

class AuthService {
  private readonly config = oauthConfig;

  /**
   * Register a new user
   */
  async register(userData: RegisterData): Promise<User> {
    try {
      const response = await apiClient.post('/auth/register', userData);
      
      // Handle different response structures
      const data = response.data || response;
      
      if (data.tokens) {
        // Store tokens if registration includes auto-login
        StorageUtils.storeTokens(
          data.tokens.access_token,
          data.tokens.refresh_token,
          data.tokens.id_token
        );
      }
      
      return data.user || data;
    } catch (error) {
      console.error('Registration failed:', error);
      const errorMessage = error instanceof Error ? error.message : 'Registration failed';
      throw new Error(errorMessage);
    }
  }

  /**
   * Login with email and password (local authentication)
   */
  async loginWithCredentials(email: string, password: string): Promise<User> {
    try {
      const loginData: LoginCredentials = { email, password };
      const response = await apiClient.post('/auth/login', loginData);

      console.log('Login response:', response);

      // Handle different response structures
      const data = response.data || response;
      const tokens = data.tokens;

      if (!tokens?.access_token) {
        throw new Error('No access token received from server');
      }

      // Store tokens
      StorageUtils.storeTokens(
        tokens.access_token,
        tokens.refresh_token,
        tokens.id_token
      );

      // Get user info from the login response or fetch it
      let user = data.user;
      if (!user) {
        user = await this.getUserProfile();
      }

      return user;
    } catch (error) {
      console.error('Login failed:', error);
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      throw new Error(errorMessage);
    }
  }

  /**
   * Get user profile using local auth token (not OAuth)
   */
  async getUserProfile(): Promise<User> {
    try {
      const response = await apiClient.get('/auth/profile');
      return response.data || response;
    } catch (error) {
      console.error('Failed to get user profile:', error);
      throw new Error('Failed to get user profile');
    }
  }

  /**
   * Start OAuth2 flow (for OAuth authentication)
   */
  async startOAuthFlow(): Promise<void> {
    try {
      // Generate PKCE parameters
      const codeVerifier = PKCEUtils.generateCodeVerifier();
      const codeChallenge = await PKCEUtils.generateCodeChallenge(codeVerifier);
      const state = PKCEUtils.generateState();

      // Store code verifier and state for later use
      StorageUtils.storeCodeVerifier(codeVerifier);
      StorageUtils.storeState(state);

      // Build authorization URL
      const authParams = {
        response_type: 'code',
        client_id: this.config.clientId,
        redirect_uri: this.config.redirectUri,
        scope: this.config.scope || 'openid profile email',
        state,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256',
      };

      const authUrl = URLUtils.buildAuthorizationURL(this.config.authorizationEndpoint, authParams);
      
      console.log('Starting OAuth flow, redirecting to:', authUrl);
      
      // Store that we're in OAuth flow before redirect
      sessionStorage.setItem('oauth_flow_started', 'true');
      window.location.href = authUrl;
    } catch (error) {
      console.error('OAuth flow initialization failed:', error);
      throw new Error('Failed to initialize OAuth flow');
    }
  }

  /**
   * OAuth login (for backward compatibility)
   */
  async login(): Promise<void> {
    await this.startOAuthFlow();
  }

  /**
   * Handle OAuth2 callback and exchange code for tokens
   */
  async handleCallback(): Promise<User> {
    try {
      const params = URLUtils.parseQueryParams();
      
      if (params.error) {
        throw new Error(`OAuth error: ${params.error} - ${params.error_description || 'Unknown error'}`);
      }

      if (!params.code || !params.state) {
        throw new Error('Missing authorization code or state parameter');
      }

      // Verify state parameter
      const storedState = StorageUtils.getState();
      if (!storedState || storedState !== params.state) {
        throw new Error('Invalid state parameter');
      }

      // Get stored code verifier
      const codeVerifier = StorageUtils.getCodeVerifier();
      if (!codeVerifier) {
        throw new Error('Missing code verifier');
      }

      // Exchange authorization code for tokens
      const tokenResponse = await this.exchangeCodeForTokens(params.code, codeVerifier);
      
      // Store tokens
      StorageUtils.storeTokens(
        tokenResponse.access_token,
        tokenResponse.refresh_token,
        tokenResponse.id_token
      );

      // Clear temporary OAuth state
      StorageUtils.clearState();
      StorageUtils.clearCodeVerifier();

      // Get user information
      const user = await this.getUserInfo();
      return user;
    } catch (error) {
      console.error('Callback handling failed:', error);
      throw error;
    }
  }

  /**
   * Exchange authorization code for tokens
   */
  private async exchangeCodeForTokens(code: string, codeVerifier: string): Promise<TokenResponse> {
    try {
      const tokenParams = {
        grant_type: 'authorization_code',
        client_id: this.config.clientId,
        code,
        redirect_uri: this.config.redirectUri,
        code_verifier: codeVerifier,
      };

      const formData = new URLSearchParams();
      Object.entries(tokenParams).forEach(([key, value]) => {
        formData.append(key, value);
      });

      const response = await fetch(this.config.tokenEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: formData.toString(),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(`Token exchange failed: ${errorData.error || response.statusText}`);
      }

      const tokenResponse: TokenResponse = await response.json();
      
      if (!tokenResponse.access_token) {
        throw new Error('Invalid token response: missing access token');
      }

      return tokenResponse;
    } catch (error) {
      console.error('Token exchange failed:', error);
      throw error;
    }
  }

  /**
   * Get current user information using OAuth access token
   */
  async getUserInfo(): Promise<User> {
    try {
      // Get the OAuth access token (not the auth token)
      const oauthAccessToken = StorageUtils.getAccessToken();
      if (!oauthAccessToken) {
        throw new Error('No OAuth access token available');
      }

      // Make request with explicit OAuth access token
      const response = await fetch(`http://localhost:8080/api/auth/oauth/userinfo`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${oauthAccessToken}`,
          'Accept': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`UserInfo request failed: ${response.status} ${response.statusText}`);
      }

      const userInfo = await response.json();
      return userInfo;
    } catch (error) {
      console.error('Failed to get user info:', error);
      throw new Error('Failed to get user information');
    }
  }

  /**
   * Refresh access token using refresh token
   */
  async refreshToken(): Promise<boolean> {
    try {
      const refreshToken = StorageUtils.getRefreshToken();
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      const formData = new URLSearchParams();
      formData.append('grant_type', 'refresh_token');
      formData.append('client_id', this.config.clientId);
      formData.append('refresh_token', refreshToken);

      const response = await fetch(this.config.tokenEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: formData.toString(),
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const tokenResponse: TokenResponse = await response.json();
      
      // Store new tokens
      StorageUtils.storeTokens(
        tokenResponse.access_token,
        tokenResponse.refresh_token || refreshToken, // Use new refresh token if provided, otherwise keep existing
        tokenResponse.id_token
      );

      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      // Clear tokens on refresh failure
      StorageUtils.clearTokens();
      return false;
    }
  }

  /**
   * Logout user and clear all stored data
   */
  async logout(): Promise<void> {
    try {
      // Clear tokens and OAuth state
      StorageUtils.clearAll();
      
      // Simple logout without endpoint redirect for now
      window.location.href = '/';
    } catch (error) {
      console.error('Logout failed:', error);
      // Even if logout endpoint fails, clear local storage
      StorageUtils.clearAll();
    }
  }

  /**
   * Check if user is currently authenticated
   */
  isAuthenticated(): boolean {
    const accessToken = StorageUtils.getAccessToken();
    if (!accessToken) {
      return false;
    }

    // Check if token is still valid
    try {
      const tokenPayload = TokenUtils.parseJWT(accessToken);
      return !TokenUtils.isTokenExpired(tokenPayload);
    } catch {
      return false;
    }
  }

  /**
   * Get current access token
   */
  getAccessToken(): string | null {
    return StorageUtils.getAccessToken();
  }

  /**
   * Verify email with token
   */
  async verifyEmail(token: string): Promise<{ user: any; message: string }> {
    try {
      const response = await apiClient.get(`/auth/verify-email?token=${token}`);
      return response;
    } catch (error) {
      throw error;
    }
  }
}

// Export singleton instance
export const authService = new AuthService();