/**
 * StreamForge specific error class
 */
export class StreamForgeError extends Error {
  public readonly code: string;
  public readonly statusCode?: number;
  public readonly originalError?: any;

  constructor(message: string, originalError?: any, code?: string, statusCode?: number) {
    super(message);
    this.name = 'StreamForgeError';
    this.code = code || 'UNKNOWN_ERROR';
    this.statusCode = statusCode;
    this.originalError = originalError;

    // Maintain proper stack trace
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, StreamForgeError);
    }
  }
}

/**
 * Validation error
 */
export class ValidationError extends StreamForgeError {
  constructor(message: string, field?: string) {
    super(message, undefined, 'VALIDATION_ERROR');
    this.name = 'ValidationError';
    if (field) {
      this.message = `${field}: ${message}`;
    }
  }
}

/**
 * Authentication error
 */
export class AuthenticationError extends StreamForgeError {
  constructor(message: string = 'Authentication failed') {
    super(message, undefined, 'AUTHENTICATION_ERROR', 401);
    this.name = 'AuthenticationError';
  }
}

/**
 * Authorization error
 */
export class AuthorizationError extends StreamForgeError {
  constructor(message: string = 'Access denied') {
    super(message, undefined, 'AUTHORIZATION_ERROR', 403);
    this.name = 'AuthorizationError';
  }
}

/**
 * Rate limit error
 */
export class RateLimitError extends StreamForgeError {
  constructor(message: string = 'Rate limit exceeded', retryAfter?: number) {
    super(message, undefined, 'RATE_LIMIT_ERROR', 429);
    this.name = 'RateLimitError';
    if (retryAfter) {
      this.message = `${message}. Retry after ${retryAfter} seconds`;
    }
  }
}

/**
 * Network error
 */
export class NetworkError extends StreamForgeError {
  constructor(message: string = 'Network error', originalError?: any) {
    super(message, originalError, 'NETWORK_ERROR');
    this.name = 'NetworkError';
  }
}

/**
 * Timeout error
 */
export class TimeoutError extends StreamForgeError {
  constructor(message: string = 'Request timeout', timeout?: number) {
    super(message, undefined, 'TIMEOUT_ERROR');
    this.name = 'TimeoutError';
    if (timeout) {
      this.message = `${message} after ${timeout}ms`;
    }
  }
} 