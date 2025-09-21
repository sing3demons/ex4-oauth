import axios, { AxiosResponse } from 'axios';
import { randomBytes, createHash } from 'crypto';
import { v4 as uuidv4 } from 'uuid';
import { config } from './config';
import { 
  OAuth2Config, 
  TokenResponse, 
  UserInfo, 
  OIDCDiscovery,
  OAuth2Error 
} from './types';

export class OAuth2Service {
  private config: OAuth2Config;
  private discoveryCache?: OIDCDiscovery;

  constructor(oauthConfig: OAuth2Config) {
    this.config = oauthConfig;
  }

  /**
   * Generate random state parameter
   */
  generateState(): string {
    return randomBytes(16).toString('base64url');
  }

  /**
   * Generate random nonce parameter  
   */
  generateNonce(): string {
    return randomBytes(16).toString('base64url');
  }

  /**
   * Get OIDC discovery document
   */
  async getDiscovery(): Promise<OIDCDiscovery> {
    if (this.discoveryCache) {
      return this.discoveryCache;
    }

    try {
      const response: AxiosResponse<OIDCDiscovery> = await axios.get(
        `${this.config.serverUrl}/api/auth/oauth/.well-known/openid-configuration`
      );
      
      this.discoveryCache = response.data;
      return this.discoveryCache;
    } catch (error) {
      throw new Error(`Failed to fetch OIDC discovery: ${error}`);
    }
  }

  /**
   * Generate authorization URL
   */
  async generateAuthorizationUrl(
    state: string, 
    nonce: string
  ): Promise<string> {
    const discovery = await this.getDiscovery();
    const authUrl = new URL(discovery.authorization_endpoint);
    
    const params = {
      response_type: 'code',
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      scope: this.config.scope || 'openid profile email',
      state,
      nonce
    };

    // Add parameters to URL
    Object.entries(params).forEach(([key, value]) => {
      if (value) {
        authUrl.searchParams.append(key, value);
      }
    });

    return authUrl.toString();
  }

  /**
   * Exchange authorization code for tokens
   */
  async exchangeCodeForTokens(code: string): Promise<TokenResponse> {
    const discovery = await this.getDiscovery();
    
    const params = new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: this.config.redirectUri,
      client_id: this.config.clientId,
      client_secret: this.config.clientSecret
    });

    try {
      const response: AxiosResponse<TokenResponse> = await axios.post(
        discovery.token_endpoint,
        params,
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );

      return response.data;
    } catch (error: any) {
      if (error.response?.data) {
        const oauthError: OAuth2Error = error.response.data;
        throw new Error(`OAuth2 Error: ${oauthError.error} - ${oauthError.error_description}`);
      }
      throw new Error(`Token exchange failed: ${error.message}`);
    }
  }

  /**
   * Get user info using access token
   */
  async getUserInfo(accessToken: string): Promise<UserInfo> {
    const discovery = await this.getDiscovery();
    
    try {
      const response: AxiosResponse<UserInfo> = await axios.get(
        discovery.userinfo_endpoint,
        {
          headers: {
            'Authorization': `Bearer ${accessToken}`
          }
        }
      );

      return response.data;
    } catch (error: any) {
      if (error.response?.status === 401) {
        throw new Error('Access token expired or invalid');
      }
      throw new Error(`Failed to get user info: ${error.message}`);
    }
  }

  /**
   * Refresh access token
   */
  async refreshToken(refreshToken: string): Promise<TokenResponse> {
    const discovery = await this.getDiscovery();
    
    const params = new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
      client_id: this.config.clientId,
      client_secret: this.config.clientSecret
    });

    try {
      const response: AxiosResponse<TokenResponse> = await axios.post(
        discovery.token_endpoint,
        params,
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );

      return response.data;
    } catch (error: any) {
      if (error.response?.data) {
        const oauthError: OAuth2Error = error.response.data;
        throw new Error(`Refresh failed: ${oauthError.error} - ${oauthError.error_description}`);
      }
      throw new Error(`Token refresh failed: ${error.message}`);
    }
  }

  /**
   * Revoke token
   */
  async revokeToken(token: string): Promise<void> {
    try {
      await axios.post(
        `${this.config.serverUrl}/api/auth/oauth/revoke`,
        new URLSearchParams({
          token,
          client_id: this.config.clientId,
          client_secret: this.config.clientSecret
        }),
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );
    } catch (error: any) {
      // Revocation might not be implemented, so we don't throw
      console.warn('Token revocation failed:', error.message);
    }
  }
}

// Create singleton instance
export const oauthService = new OAuth2Service(config.oauth);