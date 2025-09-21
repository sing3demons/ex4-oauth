import express, { Request, Response, Router } from 'express';
import { createHmac } from 'crypto';
import { oauthService } from './oauth';
import { config } from './config';

const router = Router();

// State payload interface
interface StatePayload {
  state: string;
  timestamp: number;
  client: string;
  login_url: string;
}

/**
 * Create signed state for OAuth2 flow
 */
function createSignedState(state: string, clientId: string, clientSecret: string, req: Request): string {
  const statePayload: StatePayload = {
    state,
    timestamp: Date.now(),
    client: clientId,
    login_url: `${config.app.url}${req.originalUrl}`
  };

  const stateData = JSON.stringify(statePayload);
  const signature = createHmac('sha256', clientSecret)
    .update(stateData)
    .digest('hex');

  return Buffer.from(stateData).toString('base64') + '.' + signature;
}

/**
 * Verify signed state from OAuth2 callback
 */
function verifySignedState(signedState: string, clientId: string, clientSecret: string, req: Request): boolean {
  try {
    const stateParts = signedState.split('.');

    if (stateParts.length !== 2) {
      return false;
    }

    const [stateDataB64, receivedSignature] = stateParts;
    const stateData = Buffer.from(stateDataB64, 'base64').toString();

    // Verify signature
    const expectedSignature = createHmac('sha256', clientSecret)
      .update(stateData)
      .digest('hex');

    if (receivedSignature !== expectedSignature) {
      return false;
    }

    // Parse state payload
    const statePayload: StatePayload = JSON.parse(stateData);

    // Validate timestamp (10 minutes expiry)
    const now = Date.now();
    const stateAge = now - statePayload.timestamp;

    if (stateAge > 10 * 60 * 1000) {
      return false;
    }

    // Validate client ID
    if (statePayload.client !== clientId) {
      return false;
    }

    // Validate login URL - check if it matches the expected pattern
    const expectedLoginUrl = `${config.app.url}/auth/login`;
    if (!statePayload.login_url.startsWith(config.app.url) || 
        !statePayload.login_url.includes('/auth/login')) {
      return false;
    }

    return true;
  } catch (error) {
    return false;
  }
}

/**
 * Extract Bearer token from Authorization header
 */
function extractBearerToken(req: Request): string | null {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return null;
  }
  return authHeader.split(' ')[1];
}

/**
 * Home page with authentication status
 */
router.get('/', (req: Request, res: Response) => {
  const authHeader = req.headers.authorization;
  const hasToken = authHeader && authHeader.startsWith('Bearer ');

  res.json({
    app: config.app.name,
    authenticated: hasToken || false,
    message: hasToken ? 'Token provided' : 'No token provided',
    endpoints: {
      login: '/auth/login',
      logout: '/auth/logout',
      profile: '/api/profile',
      protected: '/api/protected'
    }
  });
});

function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Start OAuth2 login flow
 */
router.get('/auth/login', async (req: Request, res: Response) => {
  try {
    const state = oauthService.generateState();
    const nonce = oauthService.generateNonce();

    // Create signed state for verification
    const signedState = createSignedState(state, config.oauth.clientId, config.oauth.clientSecret, req);

    const authUrl = await oauthService.generateAuthorizationUrl(signedState, nonce);

    if (isValidUrl(authUrl)) {
      res.redirect(authUrl);
      return;
    }

    res.json({
      authorization_url: authUrl,
      state: signedState,
      nonce,
      message: 'Redirect user to authorization_url'
    });
  } catch (error: any) {
    res.status(500).json({
      error: 'Failed to initiate login',
      message: error.message
    });
  }
});

/**
 * Handle OAuth2 callback
 */
router.get('/auth/callback', async (req: Request, res: Response) => {
  try {
    const { code, state } = req.query;

    if (!code || !state) {
      return res.status(400).json({
        error: 'Missing required parameters',
        message: 'code and state are required'
      });
    }

    // Verify signed state
    if (!verifySignedState(state as string, config.oauth.clientId, config.oauth.clientSecret, req)) {
      return res.status(400).json({
        error: 'Invalid state',
        message: 'State verification failed'
      });
    }

    // Exchange code for token
    const tokenResponse = await oauthService.exchangeCodeForTokens(code as string);

    return res.json({
      access_token: tokenResponse.access_token,
      token_type: tokenResponse.token_type,
      expires_in: tokenResponse.expires_in,
      refresh_token: tokenResponse.refresh_token,
      message: 'Login successful'
    });
  } catch (error: any) {
    return res.status(500).json({
      error: 'Failed to handle callback',
      message: error.message
    });
  }
});

/**
 * Logout endpoint
 */
router.post('/auth/logout', async (req: Request, res: Response) => {
  const authHeader = req.headers.authorization;

  if (authHeader && authHeader.startsWith('Bearer ')) {
    const token = authHeader.substring(7);

    // Optionally revoke tokens
    try {
      await oauthService.revokeToken(token);
    } catch (err) {
      console.warn('Token revocation failed:', err);
    }
  }

  res.json({ message: 'Logged out successfully' });
});

/**
 * Get user profile using Bearer token
 */
router.get('/auth/profile', async (req: Request, res: Response) => {
  try {
    const token = extractBearerToken(req);
    if (!token) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Bearer token required'
      });
    }

    const userInfo = await oauthService.getUserInfo(token);
    return res.json(userInfo);
  } catch (error: any) {
    return res.status(500).json({
      error: 'Failed to get user profile',
      message: error.message
    });
  }
});

/**
 * Protected API endpoint example
 */
router.get('/api/protected', async (req: Request, res: Response): Promise<void> => {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    res.status(401).json({ error: 'No access token provided' });
    return;
  }

  const accessToken = authHeader.substring(7);

  try {
    // Get user info first to verify token
    const userInfo = await oauthService.getUserInfo(accessToken);

    // Example: Call OAuth server's protected endpoint
    const response = await fetch(`${config.oauth.serverUrl}/api/auth/profile`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`Server responded with ${response.status}`);
    }

    const data = await response.json();

    res.json({
      message: 'Protected data accessed successfully',
      server_response: data,
      client_user: userInfo
    });

  } catch (error: any) {
    res.status(401).json({
      error: 'Invalid or expired token',
      message: error.message
    });
  }
});

/**
 * Health check endpoint
 */
router.get('/health', (req: Request, res: Response) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime()
  });
});

/**
 * API info endpoint
 */
router.get('/api/info', (req: Request, res: Response) => {
  const authHeader = req.headers.authorization;
  const hasToken = authHeader && authHeader.startsWith('Bearer ');

  res.json({
    app: config.app.name,
    version: '1.0.0',
    oauth: {
      server_url: config.oauth.serverUrl,
      client_id: config.oauth.clientId,
      redirect_uri: config.oauth.redirectUri,
      scope: config.oauth.scope
    },
    authenticated: hasToken || false
  });
});

export default router;