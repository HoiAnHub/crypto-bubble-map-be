import { Request, Response, NextFunction } from 'express';
import { logger } from '@/utils/logger';
import { ApiError } from '@/types';

export const errorHandler = (
  error: Error | ApiError,
  req: Request,
  res: Response,
  next: NextFunction
): void => {
  // Log the error
  logger.error(`Error ${req.method} ${req.path}:`, {
    error: error.message,
    stack: error.stack,
    body: req.body,
    params: req.params,
    query: req.query,
  });

  // Default error response
  let statusCode = 500;
  let message = 'Internal Server Error';
  let code = 'INTERNAL_ERROR';
  let details: any = undefined;

  // Handle specific error types
  if ('statusCode' in error && error.statusCode) {
    statusCode = error.statusCode;
    message = error.message;
    code = error.code || 'API_ERROR';
    details = error.details;
  } else if (error.name === 'ValidationError') {
    statusCode = 400;
    message = 'Validation Error';
    code = 'VALIDATION_ERROR';
    details = error.message;
  } else if (error.name === 'CastError') {
    statusCode = 400;
    message = 'Invalid ID format';
    code = 'INVALID_ID';
  } else if (error.name === 'MongoError' || error.name === 'MongoServerError') {
    statusCode = 500;
    message = 'Database Error';
    code = 'DATABASE_ERROR';
  } else if (error.message.includes('ECONNREFUSED')) {
    statusCode = 503;
    message = 'Service Unavailable';
    code = 'SERVICE_UNAVAILABLE';
  } else if (error.message.includes('timeout')) {
    statusCode = 408;
    message = 'Request Timeout';
    code = 'TIMEOUT';
  }

  // Don't expose internal errors in production
  if (process.env.NODE_ENV === 'production' && statusCode === 500) {
    message = 'Internal Server Error';
    details = undefined;
  }

  // Send error response
  res.status(statusCode).json({
    success: false,
    error: message,
    code,
    details,
    timestamp: new Date().toISOString(),
    path: req.path,
    method: req.method,
  });
};

// Custom error classes
export class ValidationError extends Error implements ApiError {
  statusCode = 400;
  code = 'VALIDATION_ERROR';
  
  constructor(message: string, public details?: any) {
    super(message);
    this.name = 'ValidationError';
  }
}

export class NotFoundError extends Error implements ApiError {
  statusCode = 404;
  code = 'NOT_FOUND';
  
  constructor(message: string = 'Resource not found') {
    super(message);
    this.name = 'NotFoundError';
  }
}

export class UnauthorizedError extends Error implements ApiError {
  statusCode = 401;
  code = 'UNAUTHORIZED';
  
  constructor(message: string = 'Unauthorized') {
    super(message);
    this.name = 'UnauthorizedError';
  }
}

export class ForbiddenError extends Error implements ApiError {
  statusCode = 403;
  code = 'FORBIDDEN';
  
  constructor(message: string = 'Forbidden') {
    super(message);
    this.name = 'ForbiddenError';
  }
}

export class ConflictError extends Error implements ApiError {
  statusCode = 409;
  code = 'CONFLICT';
  
  constructor(message: string, public details?: any) {
    super(message);
    this.name = 'ConflictError';
  }
}

export class ServiceUnavailableError extends Error implements ApiError {
  statusCode = 503;
  code = 'SERVICE_UNAVAILABLE';
  
  constructor(message: string = 'Service Unavailable') {
    super(message);
    this.name = 'ServiceUnavailableError';
  }
}

export class RateLimitError extends Error implements ApiError {
  statusCode = 429;
  code = 'RATE_LIMIT_EXCEEDED';
  
  constructor(message: string = 'Rate limit exceeded') {
    super(message);
    this.name = 'RateLimitError';
  }
}
