from flask import Flask, request, jsonify
import os
import sys
from datetime import datetime
from dotenv import load_dotenv

load_dotenv()

# DevSmith Logger (simplified from docs/integrations/python/logger.py)
import json
import time
import threading
from urllib import request as urllib_request
from urllib.error import URLError

class DevSmithLogger:
    def __init__(self, api_url, api_key, project_slug, service_name, buffer_size=100, flush_interval=5.0):
        self.api_url = api_url
        self.api_key = api_key
        self.project_slug = project_slug
        self.service_name = service_name
        self.buffer_size = buffer_size
        self.flush_interval = flush_interval
        
        self.buffer = []
        self.lock = threading.Lock()
        self.flush_timer = None
    
    def _schedule_flush(self):
        if self.flush_timer:
            self.flush_timer.cancel()
        self.flush_timer = threading.Timer(self.flush_interval, self.flush)
        self.flush_timer.daemon = True
        self.flush_timer.start()
    
    def flush(self):
        with self.lock:
            if not self.buffer:
                return
            
            batch = self.buffer[:]
            self.buffer.clear()
        
        try:
            payload = json.dumps({
                'project_slug': self.project_slug,
                'logs': batch
            }).encode('utf-8')
            
            req = urllib_request.Request(
                self.api_url,
                data=payload,
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {self.api_key}'
                },
                method='POST'
            )
            
            urllib_request.urlopen(req, timeout=5)
        except Exception as e:
            print(f'DevSmith flush error: {e}', file=sys.stderr)
    
    def _log(self, level, message, context=None, tags=None):
        log_entry = {
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'level': level,
            'message': message,
            'service': self.service_name,
            'context': context or {},
            'tags': tags or []
        }
        
        with self.lock:
            self.buffer.append(log_entry)
            
            if len(self.buffer) >= self.buffer_size:
                self.flush()
            else:
                self._schedule_flush()
    
    def debug(self, message, context=None, tags=None):
        self._log('DEBUG', message, context, tags)
    
    def info(self, message, context=None, tags=None):
        self._log('INFO', message, context, tags)
    
    def warn(self, message, context=None, tags=None):
        self._log('WARN', message, context, tags)
    
    def error(self, message, context=None, tags=None):
        self._log('ERROR', message, context, tags)

# Flask Extension (simplified from docs/integrations/python/flask_extension.py)
class DevSmithLogging:
    def __init__(self, app, logger, skip_paths=None):
        self.app = app
        self.logger = logger
        self.skip_paths = skip_paths or []
        
        app.before_request(self._before_request)
        app.after_request(self._after_request)
        
        # Register shutdown handler
        import atexit
        atexit.register(self.logger.flush)
    
    def _before_request(self):
        if request.path in self.skip_paths:
            return
        
        request._start_time = time.time()
        
        self.logger.info('Incoming request', {
            'method': request.method,
            'path': request.path,
            'headers': dict(request.headers),
            'args': dict(request.args)
        }, ['request'])
    
    def _after_request(self, response):
        if request.path in self.skip_paths:
            return response
        
        duration = int((time.time() - getattr(request, '_start_time', time.time())) * 1000)
        
        self.logger.info('Request completed', {
            'method': request.method,
            'path': request.path,
            'status_code': response.status_code,
            'duration_ms': duration
        }, ['response'])
        
        return response
    
    def log_route(self, context=None, tags=None):
        """Decorator to add custom logging to routes"""
        def decorator(f):
            def wrapped(*args, **kwargs):
                self.logger.debug(f'Executing {f.__name__}', context, tags or [])
                return f(*args, **kwargs)
            wrapped.__name__ = f.__name__
            return wrapped
        return decorator

# Initialize logger
logger = DevSmithLogger(
    api_url=os.environ['DEVSMITH_API_URL'],
    api_key=os.environ['DEVSMITH_API_KEY'],
    project_slug=os.environ['DEVSMITH_PROJECT_SLUG'],
    service_name=os.environ['DEVSMITH_SERVICE_NAME'],
    buffer_size=100,
    flush_interval=5.0
)

# Create Flask app
app = Flask(__name__)

# Add DevSmith extension (skip health checks)
devsmith_ext = DevSmithLogging(app, logger, skip_paths=['/health'])

# Routes
@app.route('/')
def root():
    logger.info('Root endpoint accessed', {
        'ip': request.remote_addr
    }, ['endpoint', 'public'])
    
    return jsonify({
        'status': 'ok',
        'message': 'DevSmith Flask Sample App',
        'endpoints': [
            'GET / - This page',
            'GET /health - Health check (not logged)',
            'GET /api/users - Get users list',
            'POST /api/users - Create user',
            'GET /api/error - Trigger error for testing'
        ]
    })

@app.route('/health')
def health():
    # Health check endpoint - skipped by extension
    return jsonify({'status': 'healthy'})

@app.route('/api/users')
@devsmith_ext.log_route(context={'endpoint': 'get_users'}, tags=['users', 'api'])
def get_users():
    logger.debug('Fetching users list', {
        'page': request.args.get('page', 1),
        'limit': request.args.get('limit', 10)
    }, ['users', 'api'])
    
    # Simulate database query
    users = [
        {'id': 1, 'name': 'Alice'},
        {'id': 2, 'name': 'Bob'}
    ]
    
    return jsonify({'users': users, 'count': len(users)})

@app.route('/api/users', methods=['POST'])
@devsmith_ext.log_route(context={'endpoint': 'create_user'}, tags=['users', 'create'])
def create_user():
    user_data = request.get_json() or {}
    
    logger.info('Creating new user', {
        'username': user_data.get('username'),
        'email': user_data.get('email')
    }, ['users', 'create'])
    
    # Simulate validation
    if not user_data.get('username'):
        logger.warn('User creation failed - missing username', {
            'provided_fields': list(user_data.keys())
        }, ['validation', 'error'])
        
        return jsonify({'error': 'Username required'}), 400
    
    # Simulate user creation
    import random
    new_user = {
        'id': random.randint(1, 10000),
        **user_data,
        'created_at': datetime.utcnow().isoformat() + 'Z'
    }
    
    logger.info('User created successfully', {
        'user_id': new_user['id'],
        'username': new_user['username']
    }, ['users', 'success'])
    
    return jsonify({'user': new_user}), 201

@app.route('/api/error')
def error_endpoint():
    logger.warn('Error endpoint called - simulating error', {
        'ip': request.remote_addr
    }, ['error', 'test'])
    
    try:
        # Simulate error
        raise Exception('Simulated database connection error')
    except Exception as e:
        logger.error('Application error occurred', {
            'error': str(e),
            'endpoint': '/api/error'
        }, ['error', 'exception'])
        
        return jsonify({
            'error': 'Internal server error',
            'message': str(e)
        }), 500

# 404 handler
@app.errorhandler(404)
def not_found(error):
    logger.warn('404 Not Found', {
        'method': request.method,
        'path': request.path,
        'ip': request.remote_addr
    }, ['404', 'routing'])
    
    return jsonify({'error': 'Not found'}), 404

# Error handler
@app.errorhandler(Exception)
def handle_error(error):
    logger.error('Unhandled error', {
        'error': str(error),
        'path': request.path,
        'method': request.method
    }, ['error', 'unhandled'])
    
    return jsonify({'error': 'Internal server error'}), 500

if __name__ == '__main__':
    port = int(os.environ.get('FLASK_PORT', 5001))
    
    logger.info('Flask server starting', {
        'port': port,
        'env': os.environ.get('FLASK_ENV', 'development')
    }, ['startup', 'server'])
    
    print(f'Server running on http://localhost:{port}')
    print('DevSmith logging enabled')
    
    try:
        app.run(host='0.0.0.0', port=port, debug=False)
    finally:
        logger.info('Server shutting down - flushing logs', {}, ['shutdown'])
        logger.flush()
