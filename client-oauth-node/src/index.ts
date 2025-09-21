import express from 'express';
import cors from 'cors';
import { config, validateConfig } from './config';
import routes from './routes';
import { 
  errorHandler, 
  requestLogger
} from './middleware';

// Validate configuration
try {
  validateConfig();
} catch (error: any) {
  console.error('Configuration Error:', error.message);
  process.exit(1);
}

// Create Express application
const app = express();

// Basic middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(requestLogger);

// CORS configuration
app.use(cors({
  origin: true,
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'Accept']
}));

// Routes
app.use('/', routes);

// Error handling
app.use(errorHandler);

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Not Found',
    message: `Route ${req.method} ${req.originalUrl} not found`,
    available_endpoints: [
      'GET /',
      'GET /auth/login',
      'GET /auth/callback',
      'POST /auth/logout',
      'GET /api/profile',
      'GET /api/protected',
      'GET /api/info',
      'GET /health'
    ]
  });
});

// Start server
function startServer(): void {
  app.listen(config.port, () => {
    console.log('🚀 NodeJS OAuth Client Server started');
    console.log(`📍 Server URL: ${config.app.url}`);
    console.log(`🔧 Environment: ${config.nodeEnv}`);
    console.log(`🔐 OAuth Server: ${config.oauth.serverUrl}`);
    console.log(`📋 Available endpoints:`);
    console.log(`   • Home: GET ${config.app.url}/`);
    console.log(`   • Login: GET ${config.app.url}/auth/login`);
    console.log(`   • Profile: GET ${config.app.url}/api/profile (protected)`);
    console.log(`   • Protected: GET ${config.app.url}/api/protected (protected)`);
    console.log(`   • Health: GET ${config.app.url}/health`);
    console.log(`   • Info: GET ${config.app.url}/api/info`);
  });
}

// Handle graceful shutdown
process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down server...');
  process.exit(0);
});

process.on('SIGTERM', () => {
  console.log('\n🛑 Server terminated');
  process.exit(0);
});

// Start the application
if (require.main === module) {
  startServer();
}

export default app;