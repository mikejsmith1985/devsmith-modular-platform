import unittest
import json
import time
from pathlib import Path

# Load test configuration
config_path = Path(__file__).parent / '.test-config.json'
with open(config_path) as f:
    test_config = json.load(f)

# Mock logger to track calls
class MockLogger:
    def __init__(self):
        self.log_calls = []
    
    def debug(self, message, context=None, tags=None):
        self.log_calls.append({
            'level': 'DEBUG',
            'message': message,
            'context': context,
            'tags': tags
        })
    
    def info(self, message, context=None, tags=None):
        self.log_calls.append({
            'level': 'INFO',
            'message': message,
            'context': context,
            'tags': tags
        })
    
    def warn(self, message, context=None, tags=None):
        self.log_calls.append({
            'level': 'WARN',
            'message': message,
            'context': context,
            'tags': tags
        })
    
    def error(self, message, context=None, tags=None):
        self.log_calls.append({
            'level': 'ERROR',
            'message': message,
            'context': context,
            'tags': tags
        })
    
    def close(self):
        pass

# Import Flask and extension
try:
    from flask import Flask, jsonify, request
    import sys
    sys.path.insert(0, str(Path(__file__).parent.parent / 'python'))
    from flask_extension import DevSmithFlask
    FLASK_AVAILABLE = True
except ImportError:
    FLASK_AVAILABLE = False


@unittest.skipUnless(FLASK_AVAILABLE, "Flask not installed")
class TestFlaskExtension(unittest.TestCase):
    
    def setUp(self):
        self.mock_logger = MockLogger()
        self.app = Flask(__name__)
        self.app.config['TESTING'] = True
        
    def test_initialization_valid(self):
        """Test extension initialization with valid logger"""
        devsmith = DevSmithFlask(self.app, self.mock_logger)
        self.assertIsNotNone(devsmith)
    
    def test_initialization_missing_logger(self):
        """Test initialization fails without logger"""
        with self.assertRaises(ValueError):
            DevSmithFlask(self.app, None)
    
    def test_request_logging(self):
        """Test incoming requests are logged"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/test')
        
        request_logs = [log for log in self.mock_logger.log_calls 
                       if 'Incoming request' in log['message']]
        
        self.assertEqual(len(request_logs), 1)
        self.assertEqual(request_logs[0]['level'], 'INFO')
        self.assertEqual(request_logs[0]['context']['method'], 'GET')
        self.assertEqual(request_logs[0]['context']['path'], '/test')
    
    def test_response_logging(self):
        """Test responses are logged"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True}), 201
        
        with self.app.test_client() as client:
            client.get('/test')
        
        response_logs = [log for log in self.mock_logger.log_calls 
                        if 'Request completed' in log['message']]
        
        self.assertEqual(len(response_logs), 1)
        self.assertEqual(response_logs[0]['context']['status_code'], 201)
        self.assertGreaterEqual(response_logs[0]['context']['duration'], 0)
    
    def test_request_timing(self):
        """Test request duration is tracked"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/slow')
        def slow_route():
            time.sleep(0.1)
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/slow')
        
        response_logs = [log for log in self.mock_logger.log_calls 
                        if 'Request completed' in log['message']]
        
        self.assertGreaterEqual(response_logs[0]['context']['duration'], 100)
        self.assertLess(response_logs[0]['context']['duration'], 200)
    
    def test_header_redaction(self):
        """Test sensitive headers are redacted"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/test', headers={
                'Authorization': 'Bearer secret-token-12345',
                'Cookie': 'session=abc123',
                'User-Agent': 'test-agent'
            })
        
        request_logs = [log for log in self.mock_logger.log_calls 
                       if 'Incoming request' in log['message']]
        
        headers = request_logs[0]['context']['headers']
        self.assertEqual(headers.get('Authorization'), '[REDACTED]')
        self.assertEqual(headers.get('Cookie'), '[REDACTED]')
        self.assertEqual(headers.get('User-Agent'), 'test-agent')
    
    def test_skip_paths(self):
        """Test health check endpoints are skipped"""
        DevSmithFlask(self.app, self.mock_logger, skip_paths=['/health', '/metrics'])
        
        @self.app.route('/health')
        def health():
            return jsonify({'ok': True})
        
        @self.app.route('/metrics')
        def metrics():
            return jsonify({'ok': True})
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/health')
            client.get('/metrics')
            client.get('/test')
        
        # Should only log /test
        paths = [log['context']['path'] for log in self.mock_logger.log_calls 
                if 'context' in log and 'path' in log['context']]
        
        self.assertIn('/test', paths)
        self.assertNotIn('/health', paths)
        self.assertNotIn('/metrics', paths)
    
    def test_error_handling(self):
        """Test errors are logged"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/error')
        def error_route():
            raise ValueError('Test error')
        
        @self.app.errorhandler(ValueError)
        def handle_error(error):
            return jsonify({'error': str(error)}), 500
        
        with self.app.test_client() as client:
            client.get('/error')
        
        error_logs = [log for log in self.mock_logger.log_calls 
                     if log['level'] == 'ERROR']
        
        self.assertGreater(len(error_logs), 0)
        self.assertIn('Test error', error_logs[0]['message'])
    
    def test_log_route_decorator(self):
        """Test @log_route decorator"""
        devsmith = DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/decorated')
        @devsmith.log_route('Test route accessed')
        def decorated_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/decorated')
        
        decorator_logs = [log for log in self.mock_logger.log_calls 
                         if 'Test route accessed' in log['message']]
        
        self.assertEqual(len(decorator_logs), 1)
        self.assertEqual(decorator_logs[0]['level'], 'INFO')
    
    def test_custom_context(self):
        """Test custom context in decorated routes"""
        devsmith = DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/context')
        @devsmith.log_route('Context test', context={'custom': 'data'})
        def context_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/context')
        
        context_logs = [log for log in self.mock_logger.log_calls 
                       if 'Context test' in log['message']]
        
        self.assertIsNotNone(context_logs[0]['context'])
        self.assertEqual(context_logs[0]['context']['custom'], 'data')
    
    def test_custom_tags(self):
        """Test custom tags in configuration"""
        DevSmithFlask(self.app, self.mock_logger, tags=['api', 'production'])
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True})
        
        with self.app.test_client() as client:
            client.get('/test')
        
        request_logs = [log for log in self.mock_logger.log_calls 
                       if 'Incoming request' in log['message']]
        
        tags = request_logs[0]['tags']
        self.assertIn('api', tags)
        self.assertIn('production', tags)
        self.assertIn('flask', tags)  # Default tag
    
    def test_exception_tracking(self):
        """Test exceptions include stack traces"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/exception')
        def exception_route():
            raise RuntimeError('Test exception with stack')
        
        @self.app.errorhandler(RuntimeError)
        def handle_runtime_error(error):
            return jsonify({'error': str(error)}), 500
        
        with self.app.test_client() as client:
            client.get('/exception')
        
        error_logs = [log for log in self.mock_logger.log_calls 
                     if log['level'] == 'ERROR']
        
        self.assertGreater(len(error_logs), 0)
        self.assertIn('stack', error_logs[0]['context'])
        self.assertIn('Test exception with stack', error_logs[0]['context']['stack'])
    
    def test_post_request(self):
        """Test POST requests are logged"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/post', methods=['POST'])
        def post_route():
            return jsonify({'received': True})
        
        with self.app.test_client() as client:
            client.post('/post', json={'data': 'test'})
        
        request_logs = [log for log in self.mock_logger.log_calls 
                       if 'Incoming request' in log['message']]
        
        self.assertEqual(request_logs[0]['context']['method'], 'POST')
    
    def test_performance_many_requests(self):
        """Test handling many requests efficiently"""
        DevSmithFlask(self.app, self.mock_logger)
        
        @self.app.route('/test')
        def test_route():
            return jsonify({'ok': True})
        
        num_requests = 100
        start_time = time.time()
        
        with self.app.test_client() as client:
            for _ in range(num_requests):
                client.get('/test')
        
        duration = time.time() - start_time
        
        # Should complete 100 requests in reasonable time (< 5 seconds)
        self.assertLess(duration, 5.0, 
                       f"100 requests took {duration:.2f}s (should be < 5s)")
        
        # Should have logged all requests
        self.assertGreaterEqual(len(self.mock_logger.log_calls), num_requests * 2)


if __name__ == '__main__':
    unittest.main()
