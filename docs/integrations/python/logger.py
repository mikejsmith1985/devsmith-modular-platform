"""
DevSmith Logger - Python Integration

Copy this file into your project and customize the configuration.

Usage:
    from logger import DevSmithLogger
    
    logger = DevSmithLogger(
        api_key=os.getenv('DEVSMITH_API_KEY'),
        api_url=os.getenv('DEVSMITH_API_URL', 'http://localhost:3000'),
        project_slug='my-project',
        service_name='api-server'
    )
    
    logger.info('User logged in', user_id=123)
    logger.error('Database error', code='ECONNREFUSED')
"""

import requests
import time
import atexit
import os
import threading
from datetime import datetime
from typing import Dict, Any, Optional

class DevSmithLogger:
    def __init__(
        self,
        api_key: str,
        api_url: str,
        project_slug: str,
        service_name: str,
        batch_size: int = 100,
        flush_interval: float = 5.0  # seconds
    ):
        # Validate required config
        if not api_key:
            raise ValueError('DevSmithLogger: api_key is required')
        if not project_slug:
            raise ValueError('DevSmithLogger: project_slug is required')
        if not service_name:
            raise ValueError('DevSmithLogger: service_name is required')
        
        self.api_key = api_key
        self.api_url = api_url.rstrip('/')
        self.project_slug = project_slug
        self.service_name = service_name
        self.batch_size = batch_size
        self.flush_interval = flush_interval
        
        self.buffer = []
        self.lock = threading.Lock()
        self.timer = None
        
        # Setup flush timer
        self._start_timer()
        
        # Flush on exit
        atexit.register(self.flush)
    
    def _start_timer(self):
        """Start periodic flush timer"""
        def flush_periodically():
            with self.lock:
                if self.buffer:
                    self.flush()
            self._start_timer()
        
        self.timer = threading.Timer(self.flush_interval, flush_periodically)
        self.timer.daemon = True
        self.timer.start()
    
    def log(self, level: str, message: str, **context):
        """Add log entry to buffer"""
        entry = {
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'level': level.upper(),
            'message': message,
            'service': self.service_name,
            'context': context
        }
        
        with self.lock:
            self.buffer.append(entry)
            
            # Flush if batch size reached
            if len(self.buffer) >= self.batch_size:
                self.flush()
    
    def flush(self):
        """Send buffered logs to DevSmith API"""
        with self.lock:
            if not self.buffer:
                return
            
            logs = self.buffer[:]
            self.buffer.clear()
        
        payload = {
            'project_slug': self.project_slug,
            'logs': logs
        }
        
        headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {self.api_key}'
        }
        
        try:
            response = requests.post(
                f'{self.api_url}/api/logs/batch',
                json=payload,
                headers=headers,
                timeout=10
            )
            
            if response.status_code not in (200, 201):
                print(f'DevSmith Logger: Failed to send logs ({response.status_code}): {response.text}')
                # Re-add logs to buffer for retry
                with self.lock:
                    self.buffer.extend(logs)
        
        except requests.exceptions.RequestException as e:
            print(f'DevSmith Logger: Network error: {e}')
            # Re-add logs to buffer for retry
            with self.lock:
                self.buffer.extend(logs)
    
    # Convenience methods
    def debug(self, message: str, **context):
        self.log('DEBUG', message, **context)
    
    def info(self, message: str, **context):
        self.log('INFO', message, **context)
    
    def warn(self, message: str, **context):
        self.log('WARN', message, **context)
    
    def error(self, message: str, **context):
        self.log('ERROR', message, **context)
