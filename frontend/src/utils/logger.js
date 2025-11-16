// Frontend error logging utility
// Sends errors to the Logs service for centralized monitoring

const LOGS_API_URL = '/api/logs';

/**
 * Log levels matching backend
 */
export const LogLevel = {
  DEBUG: 'debug',
  INFO: 'info',
  WARNING: 'warning',
  ERROR: 'error',
  CRITICAL: 'critical'
};

/**
 * Send log entry to backend Logs service
 * @param {string} level - Log level (debug, info, warning, error, critical)
 * @param {string} message - Log message
 * @param {object} metadata - Additional context (error stack, user action, etc.)
 * @param {string[]} tags - Tags for filtering
 */
export async function sendLog(level, message, metadata = {}, tags = []) {
  try {
    const logEntry = {
      service: 'frontend',
      level: level,
      message: message,
      metadata: {
        ...metadata,
        url: window.location.href,
        userAgent: navigator.userAgent,
        timestamp: new Date().toISOString()
      },
      tags: ['frontend', ...tags]
    };

    const isDebugEnabled = import.meta.env.DEV || import.meta.env.VITE_DEBUG === 'true';
    
    if (isDebugEnabled) {
      console.log('[LOGGER] Sending log to backend:', { level, message, metadata });
    }

    // Send to logs service (don't await - fire and forget)
    fetch(LOGS_API_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(logEntry)
    })
    .then(response => {
      if (!response.ok) {
        if (isDebugEnabled) {
          console.error('[LOGGER] Failed to send log:', response.status, response.statusText);
        }
        return response.text().then(text => {
          if (isDebugEnabled) {
            console.error('[LOGGER] Response body:', text);
          }
        });
      }
      if (isDebugEnabled) {
        console.log('[LOGGER] Log sent successfully');
      }
    })
    .catch(err => {
      // If logging fails, at least log to console in debug mode
      if (isDebugEnabled) {
        console.error('[LOGGER] Network error sending log to backend:', err);
      }
    });

    // Also log to console in debug mode for immediate visibility
    if (isDebugEnabled) {
      const consoleMethod = {
        'debug': 'debug',
        'info': 'info',
        'warning': 'warn',
        'error': 'error',
        'critical': 'error'
      }[level] || 'log';
      
      console[consoleMethod](`[${level.toUpperCase()}] ${message}`, metadata);
    }
  } catch (err) {
    console.error('Logger error:', err);
  }
}

/**
 * Log an error with stack trace
 */
export function logError(error, context = {}) {
  const metadata = {
    ...context,
    errorName: error.name,
    errorMessage: error.message,
    stack: error.stack,
  };
  
  sendLog(LogLevel.ERROR, error.message, metadata, ['error', 'uncaught']);
}

/**
 * Log a warning
 */
export function logWarning(message, context = {}) {
  sendLog(LogLevel.WARNING, message, context, ['warning']);
}

/**
 * Log info
 */
export function logInfo(message, context = {}) {
  sendLog(LogLevel.INFO, message, context, ['info']);
}

/**
 * Log debug info (only in development or when VITE_DEBUG=true)
 */
export function logDebug(message, context = {}) {
  const isDebugEnabled = import.meta.env.DEV || import.meta.env.VITE_DEBUG === 'true';
  if (isDebugEnabled) {
    sendLog(LogLevel.DEBUG, message, context, ['debug']);
  }
}

/**
 * Set up global error handlers
 */
export function setupGlobalErrorHandlers() {
  // Catch unhandled errors
  window.addEventListener('error', (event) => {
    logError(event.error || new Error(event.message), {
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      type: 'unhandled_error'
    });
  });

  // Catch unhandled promise rejections
  window.addEventListener('unhandledrejection', (event) => {
    const error = event.reason instanceof Error 
      ? event.reason 
      : new Error(String(event.reason));
    
    logError(error, {
      type: 'unhandled_rejection',
      promise: event.promise
    });
  });

  // Catch console.error calls
  const originalConsoleError = console.error;
  console.error = function(...args) {
    // Call original console.error
    originalConsoleError.apply(console, args);
    
    // Send to logs service
    const message = args.map(arg => 
      typeof arg === 'object' ? JSON.stringify(arg) : String(arg)
    ).join(' ');
    
    sendLog(LogLevel.ERROR, message, {
      type: 'console_error',
      args: args
    }, ['console', 'error']);
  };
}
