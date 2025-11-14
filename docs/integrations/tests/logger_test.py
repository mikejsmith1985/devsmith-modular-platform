"""
Unit tests for Python logger (logger.py)

Tests buffer management, threading, batch sending, retry logic, and cleanup.
"""

import unittest
import json
import time
import threading
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
import sys
import os

# Add parent directory to path to import logger
sys.path.insert(0, str(Path(__file__).parent.parent / 'python'))

# Load test configuration
test_config_path = Path(__file__).parent / '.test-config.json'
with open(test_config_path) as f:
    TEST_CONFIG = json.load(f)

# Mock HTTP server for testing
received_requests = []
mock_server = None
mock_server_thread = None


class MockHTTPHandler(BaseHTTPRequestHandler):
    """Mock HTTP handler to capture requests"""
    
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length).decode('utf-8')
        
        request_data = {
            'method': 'POST',
            'path': self.path,
            'headers': dict(self.headers),
            'body': json.loads(body) if body else None
        }
        received_requests.append(request_data)
        
        # Validate API key
        if self.headers.get('X-Api-Key') != TEST_CONFIG['apiKey']:
            self.send_response(401)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'error': 'Invalid API key'}).encode())
        else:
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {
                'success': True,
                'received': len(request_data['body'].get('logs', []))
            }
            self.wfile.write(json.dumps(response).encode())
    
    def log_message(self, format, *args):
        # Suppress HTTP log messages
        pass


def start_mock_server(port=8998):
    """Start mock HTTP server in background thread"""
    global mock_server, mock_server_thread
    
    mock_server = HTTPServer(('localhost', port), MockHTTPHandler)
    mock_server_thread = threading.Thread(target=mock_server.serve_forever, daemon=True)
    mock_server_thread.start()
    time.sleep(0.5)  # Give server time to start


def stop_mock_server():
    """Stop mock HTTP server"""
    global mock_server
    if mock_server:
        mock_server.shutdown()
        mock_server.server_close()


class TestPythonLogger(unittest.TestCase):
    """Test suite for Python logger"""
    
    @classmethod
    def setUpClass(cls):
        """Set up test environment once"""
        start_mock_server()
        
        # Import logger after mock server is ready
        from logger import DevSmithLogger
        cls.Logger = DevSmithLogger
    
    @classmethod
    def tearDownClass(cls):
        """Clean up after all tests"""
        stop_mock_server()
    
    def setUp(self):
        """Reset before each test"""
        global received_requests
        received_requests = []
    
    def test_initialization_valid(self):
        """Test logger initialization with valid config"""
        logger = self.Logger(
            api_key=TEST_CONFIG['apiKey'],
            api_url='http://localhost:8998',
            project_slug=TEST_CONFIG['projectSlug'],
            service_name='test-service'
        )
        
        self.assertIsNotNone(logger)
        self.assertEqual(logger.project_slug, TEST_CONFIG['projectSlug'])
        self.assertEqual(logger.service_name, 'test-service')
    
    def test_initialization_missing_api_key(self):
        """Test logger initialization fails without API key"""
        with self.assertRaises(ValueError) as context:
            self.Logger(
                api_key=None,
                api_url='http://localhost:8998',
                project_slug='test',
                service_name='test'
            )
        self.assertIn('API key is required', str(context.exception))
    
    def test_initialization_missing_project_slug(self):
        """Test logger initialization fails without project slug"""
        with self.assertRaises(ValueError) as context:
            self.Logger(
                api_key='key',
                api_url='http://localhost:8998',
                project_slug=None,
                service_name='test'
            )
        self.assertIn('Project slug is required', str(context.exception))
    
    def test_buffer_management(self):
        """Test logs are added to buffer"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test'
        )
        
        logger.info('Message 1')
        logger.info('Message 2')
        
        self.assertEqual(len(logger.buffer), 2)
    
    def test_buffer_size_limit(self):
        """Test buffer respects custom size limit"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test',
            buffer_size=5
        )
        
        for i in range(10):
            logger.info(f'Message {i}')
        
        time.sleep(1)  # Wait for flush
        self.assertLess(len(logger.buffer), 10)
    
    def test_buffer_flush_on_full(self):
        """Test buffer flushes when full"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test',
            buffer_size=3
        )
        
        logger.info('Message 1')
        logger.info('Message 2')
        logger.info('Message 3')  # Should trigger flush
        
        time.sleep(1)  # Wait for background thread
        self.assertEqual(len(received_requests), 1)
        self.assertEqual(len(received_requests[0]['body']['logs']), 3)
    
    def test_log_levels(self):
        """Test all log levels"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test'
        )
        
        logger.debug('Debug message')
        logger.info('Info message')
        logger.warn('Warning message')
        logger.error('Error message')
        
        self.assertEqual(len(logger.buffer), 4)
        self.assertEqual(logger.buffer[0]['level'], 'DEBUG')
        self.assertEqual(logger.buffer[1]['level'], 'INFO')
        self.assertEqual(logger.buffer[2]['level'], 'WARN')
        self.assertEqual(logger.buffer[3]['level'], 'ERROR')
    
    def test_context_and_tags(self):
        """Test context and tags are included"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test'
        )
        
        logger.info('Message', context={'user_id': 123}, tags=['auth', 'user'])
        
        log_entry = logger.buffer[0]
        self.assertEqual(log_entry['context'], {'user_id': 123})
        self.assertEqual(log_entry['tags'], ['auth', 'user'])
    
    def test_batch_format(self):
        """Test batch request format is correct"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test-service',
            buffer_size=2
        )
        
        logger.info('Message 1')
        logger.info('Message 2')
        
        time.sleep(1)  # Wait for flush
        
        self.assertEqual(len(received_requests), 1)
        request = received_requests[0]
        
        self.assertEqual(request['method'], 'POST')
        self.assertEqual(request['headers']['X-Api-Key'], TEST_CONFIG['apiKey'])
        self.assertEqual(request['body']['project_slug'], TEST_CONFIG['projectSlug'])
        self.assertEqual(len(request['body']['logs']), 2)
    
    def test_batch_required_fields(self):
        """Test batch includes all required fields"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test-service',
            buffer_size=1
        )
        
        logger.info('Test message', context={'key': 'value'}, tags=['tag1'])
        
        time.sleep(1)  # Wait for flush
        
        log_entry = received_requests[0]['body']['logs'][0]
        
        self.assertIn('timestamp', log_entry)
        self.assertEqual(log_entry['level'], 'INFO')
        self.assertEqual(log_entry['message'], 'Test message')
        self.assertEqual(log_entry['service'], 'test-service')
        self.assertEqual(log_entry['context'], {'key': 'value'})
        self.assertEqual(log_entry['tags'], ['tag1'])
    
    def test_time_based_flush(self):
        """Test buffer flushes after interval"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test',
            flush_interval=2  # 2 seconds
        )
        
        logger.info('Message 1')
        
        # Should not have sent yet
        self.assertEqual(len(received_requests), 0)
        
        # Wait for flush interval
        time.sleep(2.5)
        self.assertEqual(len(received_requests), 1)
    
    def test_cleanup_on_close(self):
        """Test buffer flushes on close"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test'
        )
        
        logger.info('Message 1')
        logger.info('Message 2')
        
        logger.close()
        
        time.sleep(0.5)  # Wait for flush
        self.assertEqual(len(received_requests), 1)
        self.assertEqual(len(received_requests[0]['body']['logs']), 2)
    
    def test_threading_safety(self):
        """Test logger is thread-safe"""
        logger = self.Logger(
            TEST_CONFIG['apiKey'],
            'http://localhost:8998',
            TEST_CONFIG['projectSlug'],
            'test',
            buffer_size=100
        )
        
        def log_messages(thread_id, count):
            for i in range(count):
                logger.info(f'Thread {thread_id} - Message {i}')
        
        threads = []
        for i in range(5):
            t = threading.Thread(target=log_messages, args=(i, 10))
            threads.append(t)
            t.start()
        
        for t in threads:
            t.join()
        
        # Should have 50 logs total (5 threads x 10 messages)
        logger.close()
        time.sleep(1)
        
        total_logs = sum(len(req['body']['logs']) for req in received_requests)
        self.assertEqual(total_logs, 50)


if __name__ == '__main__':
    unittest.main(verbosity=2)
