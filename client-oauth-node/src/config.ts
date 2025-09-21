import dotenv from 'dotenv';
import { AppConfig } from './types';

// Load environment variables
dotenv.config();

export const config: AppConfig = {
  port: parseInt(process.env.PORT || '3001', 10),
  nodeEnv: process.env.NODE_ENV || 'development',
  
  oauth: {
    clientId: process.env.OAUTH_CLIENT_ID || '',
    clientSecret: process.env.OAUTH_CLIENT_SECRET || '',
    serverUrl: process.env.OAUTH_SERVER_URL || 'http://localhost:8080',
    redirectUri: process.env.OAUTH_REDIRECT_URI || 'http://localhost:3001/auth/callback',
    scope: 'openid profile email read'
  },
  
  app: {
    name: process.env.APP_NAME || 'NodeJS OAuth Client',
    url: process.env.APP_URL || 'http://localhost:3001'
  }
};

// Validate required configuration
export function validateConfig(): void {
  const required = [
    'OAUTH_CLIENT_ID',
    'OAUTH_CLIENT_SECRET'
  ];

  const missing = required.filter(key => !process.env[key]);
  
  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
  }
}

export default config;