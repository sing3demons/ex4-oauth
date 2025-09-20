import CryptoJS from 'crypto-js';

/**
 * PKCE (Proof Key for Code Exchange) Utilities
 * สำหรับ OAuth2 Authorization Code Flow with PKCE
 */
export class PKCEUtils {
  /**
   * สร้าง code_verifier (43-128 characters)
   * Base64URL-encoded string ที่ใช้สำหรับ PKCE
   */
  static generateCodeVerifier(): string {
    const array = new Uint8Array(32);
    window.crypto.getRandomValues(array);
    return btoa(String.fromCharCode.apply(null, Array.from(array)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  /**
   * สร้าง code_challenge จาก code_verifier
   * SHA256 hash ของ code_verifier แล้ว encode เป็น Base64URL
   */
  static async generateCodeChallenge(verifier: string): Promise<string> {
    const hash = CryptoJS.SHA256(verifier);
    return hash.toString(CryptoJS.enc.Base64url);
  }

  /**
   * สร้าง random state สำหรับ CSRF protection
   * ป้องกัน Cross-Site Request Forgery attacks
   */
  static generateState(): string {
    const array = new Uint8Array(16);
    window.crypto.getRandomValues(array);
    return btoa(String.fromCharCode.apply(null, Array.from(array)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  /**
   * สร้าง nonce สำหรับ ID Token validation
   * ป้องกัน replay attacks ใน OpenID Connect
   */
  static generateNonce(): string {
    return this.generateState(); // ใช้วิธีเดียวกัน
  }

  /**
   * Validate code_verifier format
   * ตรวจสอบว่า code_verifier ถูกต้องตาม RFC 7636
   */
  static validateCodeVerifier(verifier: string): boolean {
    const regex = /^[A-Za-z0-9\-_]{43,128}$/;
    return regex.test(verifier);
  }

  /**
   * Validate state parameter format
   * ตรวจสอบว่า state parameter ถูกต้อง
   */
  static validateState(state: string): boolean {
    const regex = /^[A-Za-z0-9\-_]+$/;
    return regex.test(state) && state.length >= 16;
  }
}

/**
 * Utility functions สำหรับ token management
 */
export class TokenUtils {
  /**
   * Parse JWT token (ไม่ verify signature)
   * ใช้สำหรับอ่าน claims จาก token
   */
  static parseJWT(token: string): any {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      return JSON.parse(jsonPayload);
    } catch (error) {
      console.error('Failed to parse JWT:', error);
      return null;
    }
  }

  /**
   * Check if token is expired
   * ตรวจสอบว่า token หมดอายุหรือยัง
   */
  static isTokenExpired(token: string): boolean {
    const payload = this.parseJWT(token);
    if (!payload || !payload.exp) return true;
    
    const now = Math.floor(Date.now() / 1000);
    return payload.exp < now;
  }

  /**
   * Get token expiration time
   * ดึงเวลาหมดอายุของ token
   */
  static getTokenExpiration(token: string): Date | null {
    const payload = this.parseJWT(token);
    if (!payload || !payload.exp) return null;
    
    return new Date(payload.exp * 1000);
  }

  /**
   * Check if token will expire soon (within minutes)
   * ตรวจสอบว่า token จะหมดอายุเร็ว ๆ นี้
   */
  static willExpireSoon(token: string, minutesBefore: number = 5): boolean {
    const expiration = this.getTokenExpiration(token);
    if (!expiration) return true;
    
    const now = new Date();
    const warningTime = new Date(expiration.getTime() - (minutesBefore * 60 * 1000));
    
    return now >= warningTime;
  }
}

/**
 * Storage utilities สำหรับ secure token storage
 */
export class StorageUtils {
  private static readonly ACCESS_TOKEN_KEY = 'oauth_access_token';
  private static readonly REFRESH_TOKEN_KEY = 'oauth_refresh_token';
  private static readonly ID_TOKEN_KEY = 'oauth_id_token';
  private static readonly STATE_KEY = 'oauth_state';
  private static readonly CODE_VERIFIER_KEY = 'oauth_code_verifier';

  static storeTokens(accessToken: string, refreshToken?: string, idToken?: string): void {
    localStorage.setItem(this.ACCESS_TOKEN_KEY, accessToken);
    if (refreshToken) {
      localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
    }
    if (idToken) {
      localStorage.setItem(this.ID_TOKEN_KEY, idToken);
    }
  }

  static getAccessToken(): string | null {
    return localStorage.getItem(this.ACCESS_TOKEN_KEY);
  }

  static getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  static getIdToken(): string | null {
    return localStorage.getItem(this.ID_TOKEN_KEY);
  }

  static clearTokens(): void {
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
    localStorage.removeItem(this.ID_TOKEN_KEY);
  }

  static storeState(state: string): void {
    sessionStorage.setItem(this.STATE_KEY, state);
  }

  static getState(): string | null {
    return sessionStorage.getItem(this.STATE_KEY);
  }

  static clearState(): void {
    sessionStorage.removeItem(this.STATE_KEY);
  }

  static storeCodeVerifier(verifier: string): void {
    sessionStorage.setItem(this.CODE_VERIFIER_KEY, verifier);
  }

  static getCodeVerifier(): string | null {
    return sessionStorage.getItem(this.CODE_VERIFIER_KEY);
  }

  static clearCodeVerifier(): void {
    sessionStorage.removeItem(this.CODE_VERIFIER_KEY);
  }

  static clearAll(): void {
    this.clearTokens();
    this.clearState();
    this.clearCodeVerifier();
  }
}

/**
 * URL utilities สำหรับ OAuth2 flow
 */
export class URLUtils {
  /**
   * Build authorization URL with parameters
   */
  static buildAuthorizationURL(baseURL: string, params: Record<string, string>): string {
    const url = new URL(baseURL);
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.append(key, value);
    });
    return url.toString();
  }

  /**
   * Parse query parameters from current URL
   */
  static parseQueryParams(): Record<string, string> {
    const params = new URLSearchParams(window.location.search);
    const result: Record<string, string> = {};
    
    params.forEach((value, key) => {
      result[key] = value;
    });
    
    return result;
  }

  /**
   * Clear query parameters from URL
   */
  static clearQueryParams(): void {
    const url = new URL(window.location.href);
    url.search = '';
    window.history.replaceState({}, '', url.toString());
  }

  /**
   * Get current origin (protocol + host)
   */
  static getCurrentOrigin(): string {
    return window.location.origin;
  }
}