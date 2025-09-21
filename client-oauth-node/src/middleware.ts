import { Request, Response, NextFunction } from 'express';

/**
 * Error handling middleware
 */
export function errorHandler(
  error: Error, 
  req: Request, 
  res: Response, 
  next: NextFunction
): void {
  console.error('Error:', error);
  
  const isDev = process.env.NODE_ENV === 'development';
  
  res.status(500).json({
    error: 'Internal Server Error',
    message: isDev ? error.message : 'Something went wrong',
    ...(isDev && { stack: error.stack })
  });
}

/**
 * Logging middleware
 */
export function requestLogger(req: Request, res: Response, next: NextFunction): void {
  const start = Date.now();
  
  res.on('finish', () => {
    const duration = Date.now() - start;
    console.log(`${req.method} ${req.path} - ${res.statusCode} (${duration}ms)`);
  });
  
  next();
}