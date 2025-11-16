"""
Flask Integration for DevSmith Logging

Automatically logs HTTP requests/responses to DevSmith platform.

Installation:
1. Copy logger.py into your project
2. Copy this file (flask_integration.py) into your project
3. Add to your Flask app

Usage:
    from flask import Flask
    from flask_integration import DevSmithFlask
    
    app = Flask(__name__)
    
    devsmith = DevSmithFlask(app, {
        'api_key': os.getenv('DEVSMITH_API_KEY'),
        'api_url': os.getenv('DEVSMITH_API_URL', 'http://localhost:3000'),
        'project_slug': 'my-app',
        'service_name': 'flask-api'
    })
    
    # Your routes...
    @app.route('/')
    def hello():
        return 'Hello World'
    
    if __name__ == '__main__':
        app.run()
"""

from flask import request, g
from functools import wraps
import time
import traceback
from logger import DevSmithLogger


class DevSmithFlask:
    """
    Flask extension for DevSmith logging
    
    Automatically logs all HTTP requests/responses with timing, status, and context.
    """
    
    def __init__(self, app=None, config=None):
        """
        Initialize DevSmith Flask extension
        
        Args:
            app: Flask application instance (optional)
            config: Configuration dict with keys:
                - api_key: DevSmith API key (required)
                - api_url: DevSmith API URL (default: http://localhost:3000)
                - project_slug: Project slug in DevSmith (required)
                - service_name: Service name for this app (required)
                - log_body: Log request/response bodies (default: False)
                - skip_paths: List of paths to skip (default: ['/health', '/metrics'])
                - redact_headers: List of headers to redact (default: ['authorization', 'cookie'])
        """
        self.logger = None
        self.log_body = False
        self.skip_paths = ['/health', '/metrics']
        self.redact_headers = ['authorization', 'cookie', 'x-api-key']
        
        if app is not None:
            self.init_app(app, config)
    
    def init_app(self, app, config=None):
        """
        Initialize extension with Flask app
        
        Args:
            app: Flask application instance
            config: Configuration dict
        """
        if config is None:
            config = {}
        
        # Create logger
        self.logger = DevSmithLogger(
            api_key=config.get('api_key'),
            api_url=config.get('api_url', 'http://localhost:3000'),
            project_slug=config.get('project_slug'),
            service_name=config.get('service_name')
        )
        
        # Configuration
        self.log_body = config.get('log_body', False)
        self.skip_paths = config.get('skip_paths', self.skip_paths)
        self.redact_headers = config.get('redact_headers', self.redact_headers)
        
        # Register hooks
        app.before_request(self._before_request)
        app.after_request(self._after_request)
        app.teardown_request(self._teardown_request)
        
        # Store extension on app
        if not hasattr(app, 'extensions'):
            app.extensions = {}
        app.extensions['devsmith'] = self
    
    def _before_request(self):
        """Record request start time"""
        g.devsmith_start_time = time.time()
    
    def _after_request(self, response):
        """Log request/response after handling"""
        # Skip configured paths
        if request.path in self.skip_paths:
            return response
        
        # Calculate duration
        duration_ms = int((time.time() - g.get('devsmith_start_time', time.time())) * 1000)
        
        # Determine log level based on status code
        if response.status_code >= 500:
            level = 'error'
        elif response.status_code >= 400:
            level = 'warn'
        else:
            level = 'info'
        
        # Build context
        context = {
            # Request info
            'method': request.method,
            'path': request.path,
            'query': dict(request.args),
            'ip': request.remote_addr,
            'user_agent': request.headers.get('User-Agent'),
            
            # Response info
            'status_code': response.status_code,
            'duration': f'{duration_ms}ms',
            
            # Headers (redacted)
            'request_headers': self._redact_headers(dict(request.headers)),
            'response_headers': self._redact_headers(dict(response.headers))
        }
        
        # Optionally include bodies
        if self.log_body:
            if request.data:
                context['request_body'] = request.data.decode('utf-8')[:1000]  # Limit to 1KB
            if response.data:
                context['response_body'] = response.data.decode('utf-8')[:1000]
        
        # Log message
        message = f'{request.method} {request.path} {response.status_code} {duration_ms}ms'
        
        # Log with appropriate level
        if level == 'error':
            self.logger.error(message, **context)
        elif level == 'warn':
            self.logger.warn(message, **context)
        else:
            self.logger.info(message, **context)
        
        return response
    
    def _teardown_request(self, exception=None):
        """Log unhandled exceptions"""
        if exception is not None:
            self.logger.error('Unhandled exception', **{
                'method': request.method,
                'path': request.path,
                'exception': str(exception),
                'traceback': traceback.format_exc()
            })
    
    def _redact_headers(self, headers):
        """Redact sensitive headers"""
        redacted = {}
        for key, value in headers.items():
            if key.lower() in self.redact_headers:
                redacted[key] = '[REDACTED]'
            else:
                redacted[key] = value
        return redacted


def log_route(logger, level='info'):
    """
    Decorator for logging specific routes
    
    Usage:
        @app.route('/api/users')
        @log_route(devsmith.logger, level='info')
        def get_users():
            return jsonify(users)
    
    Args:
        logger: DevSmithLogger instance
        level: Log level (debug, info, warn, error)
    """
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            start_time = time.time()
            
            try:
                result = f(*args, **kwargs)
                duration_ms = int((time.time() - start_time) * 1000)
                
                getattr(logger, level)(
                    f'Route {request.endpoint} completed',
                    method=request.method,
                    path=request.path,
                    duration=f'{duration_ms}ms'
                )
                
                return result
            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)
                
                logger.error(
                    f'Route {request.endpoint} failed',
                    method=request.method,
                    path=request.path,
                    duration=f'{duration_ms}ms',
                    error=str(e),
                    traceback=traceback.format_exc()
                )
                
                raise
        
        return decorated_function
    return decorator


# Example usage
if __name__ == '__main__':
    import os
    from flask import Flask, jsonify
    
    app = Flask(__name__)
    
    # Initialize DevSmith
    devsmith = DevSmithFlask(app, {
        'api_key': os.getenv('DEVSMITH_API_KEY'),
        'api_url': os.getenv('DEVSMITH_API_URL', 'http://localhost:3000'),
        'project_slug': 'test-project',
        'service_name': 'flask-api',
        'log_body': False,  # Set to True to log request/response bodies
        'skip_paths': ['/health', '/metrics']
    })
    
    @app.route('/')
    def hello():
        return 'Hello World'
    
    @app.route('/api/users')
    @log_route(devsmith.logger, level='info')
    def get_users():
        # This route has extra logging via decorator
        return jsonify([
            {'id': 1, 'name': 'Alice'},
            {'id': 2, 'name': 'Bob'}
        ])
    
    @app.route('/api/error')
    def error_route():
        # This will trigger error logging
        raise ValueError('Intentional error for testing')
    
    app.run(debug=True, port=5000)
