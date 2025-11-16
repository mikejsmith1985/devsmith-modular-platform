import React, { useState } from 'react';
import { Link } from 'react-router-dom';

export default function IntegrationDocsPage() {
  const [activeTab, setActiveTab] = useState('javascript');
  const [copiedCode, setCopiedCode] = useState(null);

  const handleCopyCode = (code, identifier) => {
    navigator.clipboard.writeText(code).then(() => {
      setCopiedCode(identifier);
      setTimeout(() => setCopiedCode(null), 2000);
    });
  };

  const languages = {
    javascript: {
      label: 'JavaScript / Node.js',
      icon: 'bi-filetype-js',
      samples: [
        {
          title: 'Basic Setup',
          description: 'Minimal integration for Node.js applications',
          code: `// logs-client.js
const LOGS_API_URL = process.env.LOGS_API_URL || 'http://localhost:8082';
const LOGS_API_KEY = process.env.LOGS_API_KEY; // Get from Projects page

const logBatch = [];
const MAX_BATCH_SIZE = 1000;
const FLUSH_INTERVAL = 5000; // 5 seconds

async function sendBatch(entries) {
  if (entries.length === 0) return;
  
  try {
    const response = await fetch(\`\${LOGS_API_URL}/api/logs/batch\`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': \`Bearer \${LOGS_API_KEY}\`
      },
      body: JSON.stringify({ entries })
    });
    
    if (!response.ok) {
      console.error('Failed to send logs:', response.status);
    }
  } catch (error) {
    console.error('Error sending logs:', error);
  }
}

function log(level, message, metadata = {}) {
  logBatch.push({
    level,
    message,
    service: process.env.SERVICE_NAME || 'my-app',
    metadata,
    timestamp: new Date().toISOString()
  });
  
  if (logBatch.length >= MAX_BATCH_SIZE) {
    sendBatch([...logBatch]);
    logBatch.length = 0;
  }
}

// Flush logs periodically
setInterval(() => {
  if (logBatch.length > 0) {
    sendBatch([...logBatch]);
    logBatch.length = 0;
  }
}, FLUSH_INTERVAL);

// Export logging functions
module.exports = {
  debug: (msg, meta) => log('DEBUG', msg, meta),
  info: (msg, meta) => log('INFO', msg, meta),
  warn: (msg, meta) => log('WARNING', msg, meta),
  error: (msg, meta) => log('ERROR', msg, meta)
};`,
          language: 'javascript'
        },
        {
          title: 'Express.js Middleware',
          description: 'Automatic request logging for Express applications',
          code: `// middleware/logging.js
const logger = require('./logs-client');

function loggingMiddleware(req, res, next) {
  const start = Date.now();
  
  // Log request
  logger.info('HTTP Request', {
    method: req.method,
    path: req.path,
    query: req.query,
    ip: req.ip
  });
  
  // Capture response
  res.on('finish', () => {
    const duration = Date.now() - start;
    const level = res.statusCode >= 500 ? 'ERROR' : 
                  res.statusCode >= 400 ? 'WARNING' : 'INFO';
    
    logger[level.toLowerCase()]('HTTP Response', {
      method: req.method,
      path: req.path,
      status: res.statusCode,
      duration_ms: duration
    });
  });
  
  next();
}

module.exports = loggingMiddleware;

// Usage in app.js:
// const loggingMiddleware = require('./middleware/logging');
// app.use(loggingMiddleware);`,
          language: 'javascript'
        }
      ]
    },
    python: {
      label: 'Python',
      icon: 'bi-filetype-py',
      samples: [
        {
          title: 'Basic Setup',
          description: 'Minimal integration for Python applications',
          code: `# logs_client.py
import os
import json
import time
import requests
from threading import Thread, Lock
from datetime import datetime

LOGS_API_URL = os.getenv('LOGS_API_URL', 'http://localhost:8082')
LOGS_API_KEY = os.getenv('LOGS_API_KEY')  # Get from Projects page
SERVICE_NAME = os.getenv('SERVICE_NAME', 'my-app')

MAX_BATCH_SIZE = 1000
FLUSH_INTERVAL = 5  # seconds

class LogsClient:
    def __init__(self):
        self.batch = []
        self.lock = Lock()
        self.running = True
        self.thread = Thread(target=self._flush_worker, daemon=True)
        self.thread.start()
    
    def _send_batch(self, entries):
        if not entries:
            return
        
        try:
            response = requests.post(
                f'{LOGS_API_URL}/api/logs/batch',
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {LOGS_API_KEY}'
                },
                json={'entries': entries},
                timeout=10
            )
            response.raise_for_status()
        except Exception as e:
            print(f'Error sending logs: {e}')
    
    def _flush_worker(self):
        while self.running:
            time.sleep(FLUSH_INTERVAL)
            self.flush()
    
    def flush(self):
        with self.lock:
            if self.batch:
                self._send_batch(self.batch[:])
                self.batch.clear()
    
    def log(self, level, message, **metadata):
        entry = {
            'level': level,
            'message': message,
            'service': SERVICE_NAME,
            'metadata': metadata,
            'timestamp': datetime.utcnow().isoformat() + 'Z'
        }
        
        with self.lock:
            self.batch.append(entry)
            if len(self.batch) >= MAX_BATCH_SIZE:
                self._send_batch(self.batch[:])
                self.batch.clear()
    
    def debug(self, message, **metadata):
        self.log('DEBUG', message, **metadata)
    
    def info(self, message, **metadata):
        self.log('INFO', message, **metadata)
    
    def warn(self, message, **metadata):
        self.log('WARNING', message, **metadata)
    
    def error(self, message, **metadata):
        self.log('ERROR', message, **metadata)

# Create global logger instance
logger = LogsClient()`,
          language: 'python'
        },
        {
          title: 'Flask Integration',
          description: 'Automatic request logging for Flask applications',
          code: `# app.py (Flask integration)
from flask import Flask, request, g
from logs_client import logger
import time

app = Flask(__name__)

@app.before_request
def log_request():
    g.start_time = time.time()
    logger.info('HTTP Request', 
        method=request.method,
        path=request.path,
        query=dict(request.args),
        ip=request.remote_addr
    )

@app.after_request
def log_response(response):
    duration = (time.time() - g.start_time) * 1000  # ms
    
    if response.status_code >= 500:
        level = 'error'
    elif response.status_code >= 400:
        level = 'warn'
    else:
        level = 'info'
    
    getattr(logger, level)('HTTP Response',
        method=request.method,
        path=request.path,
        status=response.status_code,
        duration_ms=duration
    )
    
    return response

# Your routes here
@app.route('/')
def index():
    logger.info('Index page accessed')
    return 'Hello World'

if __name__ == '__main__':
    app.run()`,
          language: 'python'
        }
      ]
    },
    go: {
      label: 'Go',
      icon: 'bi-filetype-go',
      samples: [
        {
          title: 'Basic Setup',
          description: 'Minimal integration for Go applications',
          code: `// logs_client.go
package logs

import (
\t"bytes"
\t"encoding/json"
\t"fmt"
\t"net/http"
\t"os"
\t"sync"
\t"time"
)

const (
\tMaxBatchSize  = 1000
\tFlushInterval = 5 * time.Second
)

var (
\tlogsAPIURL = os.Getenv("LOGS_API_URL")
\tlogsAPIKey = os.Getenv("LOGS_API_KEY")
\tserviceName = os.Getenv("SERVICE_NAME")
)

type LogEntry struct {
\tLevel     string                 \`json:"level"\`
\tMessage   string                 \`json:"message"\`
\tService   string                 \`json:"service"\`
\tMetadata  map[string]interface{} \`json:"metadata,omitempty"\`
\tTimestamp string                 \`json:"timestamp"\`
}

type LogsClient struct {
\tbatch      []LogEntry
\tmu         sync.Mutex
\thttpClient *http.Client
}

func NewLogsClient() *LogsClient {
\tif logsAPIURL == "" {
\t\tlogsAPIURL = "http://localhost:8082"
\t}
\tif serviceName == "" {
\t\tserviceName = "my-app"
\t}
\t
\tclient := &LogsClient{
\t\tbatch:      make([]LogEntry, 0, MaxBatchSize),
\t\thttpClient: &http.Client{Timeout: 10 * time.Second},
\t}
\t
\t// Start flush worker
\tgo client.flushWorker()
\t
\treturn client
}

func (c *LogsClient) sendBatch(entries []LogEntry) error {
\tif len(entries) == 0 {
\t\treturn nil
\t}
\t
\tbody, err := json.Marshal(map[string]interface{}{
\t\t"entries": entries,
\t})
\tif err != nil {
\t\treturn fmt.Errorf("marshal error: %w", err)
\t}
\t
\treq, err := http.NewRequest("POST", logsAPIURL+"/api/logs/batch", bytes.NewReader(body))
\tif err != nil {
\t\treturn fmt.Errorf("request creation error: %w", err)
\t}
\t
\treq.Header.Set("Content-Type", "application/json")
\treq.Header.Set("Authorization", "Bearer "+logsAPIKey)
\t
\tresp, err := c.httpClient.Do(req)
\tif err != nil {
\t\treturn fmt.Errorf("request error: %w", err)
\t}
\tdefer resp.Body.Close()
\t
\tif resp.StatusCode != http.StatusOK {
\t\treturn fmt.Errorf("unexpected status: %d", resp.StatusCode)
\t}
\t
\treturn nil
}

func (c *LogsClient) flushWorker() {
\tticker := time.NewTicker(FlushInterval)
\tdefer ticker.Stop()
\t
\tfor range ticker.C {
\t\tc.Flush()
\t}
}

func (c *LogsClient) Flush() {
\tc.mu.Lock()
\tif len(c.batch) == 0 {
\t\tc.mu.Unlock()
\t\treturn
\t}
\t
\ttoSend := make([]LogEntry, len(c.batch))
\tcopy(toSend, c.batch)
\tc.batch = c.batch[:0]
\tc.mu.Unlock()
\t
\tif err := c.sendBatch(toSend); err != nil {
\t\tfmt.Printf("Error sending logs: %v\\n", err)
\t}
}

func (c *LogsClient) log(level, message string, metadata map[string]interface{}) {
\tentry := LogEntry{
\t\tLevel:     level,
\t\tMessage:   message,
\t\tService:   serviceName,
\t\tMetadata:  metadata,
\t\tTimestamp: time.Now().UTC().Format(time.RFC3339),
\t}
\t
\tc.mu.Lock()
\tc.batch = append(c.batch, entry)
\tshouldFlush := len(c.batch) >= MaxBatchSize
\tc.mu.Unlock()
\t
\tif shouldFlush {
\t\tc.Flush()
\t}
}

func (c *LogsClient) Debug(message string, metadata map[string]interface{}) {
\tc.log("DEBUG", message, metadata)
}

func (c *LogsClient) Info(message string, metadata map[string]interface{}) {
\tc.log("INFO", message, metadata)
}

func (c *LogsClient) Warn(message string, metadata map[string]interface{}) {
\tc.log("WARNING", message, metadata)
}

func (c *LogsClient) Error(message string, metadata map[string]interface{}) {
\tc.log("ERROR", message, metadata)
}

// Global logger instance
var Logger = NewLogsClient()`,
          language: 'go'
        },
        {
          title: 'Gin Middleware',
          description: 'Automatic request logging for Gin applications',
          code: `// middleware/logging.go
package middleware

import (
\t"time"
\t"your-project/logs"
\t
\t"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\tstart := time.Now()
\t\t
\t\t// Log request
\t\tlogs.Logger.Info("HTTP Request", map[string]interface{}{
\t\t\t"method": c.Request.Method,
\t\t\t"path":   c.Request.URL.Path,
\t\t\t"query":  c.Request.URL.RawQuery,
\t\t\t"ip":     c.ClientIP(),
\t\t})
\t\t
\t\t// Process request
\t\tc.Next()
\t\t
\t\t// Log response
\t\tduration := time.Since(start).Milliseconds()
\t\tstatus := c.Writer.Status()
\t\t
\t\tvar level string
\t\tswitch {
\t\tcase status >= 500:
\t\t\tlevel = "ERROR"
\t\tcase status >= 400:
\t\t\tlevel = "WARNING"
\t\tdefault:
\t\t\tlevel = "INFO"
\t\t}
\t\t
\t\tmeta := map[string]interface{}{
\t\t\t"method":      c.Request.Method,
\t\t\t"path":        c.Request.URL.Path,
\t\t\t"status":      status,
\t\t\t"duration_ms": duration,
\t\t}
\t\t
\t\tswitch level {
\t\tcase "ERROR":
\t\t\tlogs.Logger.Error("HTTP Response", meta)
\t\tcase "WARNING":
\t\t\tlogs.Logger.Warn("HTTP Response", meta)
\t\tdefault:
\t\t\tlogs.Logger.Info("HTTP Response", meta)
\t\t}
\t}
}

// Usage in main.go:
// r := gin.Default()
// r.Use(middleware.LoggingMiddleware())`,
          language: 'go'
        }
      ]
    }
  };

  const setupSteps = [
    {
      number: 1,
      title: 'Create a Project',
      description: 'Go to the Projects page and create a new project for your application.',
      link: '/projects'
    },
    {
      number: 2,
      title: 'Copy API Key',
      description: 'Copy the generated API key (shown only once). Store it securely as an environment variable.'
    },
    {
      number: 3,
      title: 'Download Sample Code',
      description: 'Select your language below and copy the sample code. Customize service name and metadata as needed.'
    },
    {
      number: 4,
      title: 'Configure Environment',
      description: 'Set environment variables: LOGS_API_URL, LOGS_API_KEY, SERVICE_NAME'
    },
    {
      number: 5,
      title: 'Deploy & Verify',
      description: 'Deploy your application and check the Health Dashboard to see logs appearing in real-time.'
    }
  ];

  return (
    <div className="container-fluid py-4">
      {/* Header */}
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h2 className="mb-1">
            <i className="bi bi-book me-2"></i>
            Integration Documentation
          </h2>
          <p className="text-muted mb-0">
            Copy-paste examples for integrating external applications with the Logs service
          </p>
        </div>
        <Link to="/projects" className="btn btn-outline-primary">
          <i className="bi bi-folder me-2"></i>
          Manage Projects
        </Link>
      </div>

      {/* Setup Steps */}
      <div className="card mb-4">
        <div className="card-body">
          <h5 className="card-title mb-4">
            <i className="bi bi-list-ol me-2"></i>
            Quick Setup Guide
          </h5>
          <div className="row">
            {setupSteps.map((step) => (
              <div key={step.number} className="col-md-4 mb-3">
                <div className="d-flex">
                  <div className="flex-shrink-0">
                    <div className="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center" 
                         style={{ width: '32px', height: '32px', fontSize: '14px' }}>
                      {step.number}
                    </div>
                  </div>
                  <div className="flex-grow-1 ms-3">
                    <h6 className="mb-1">{step.title}</h6>
                    <p className="text-muted small mb-0">{step.description}</p>
                    {step.link && (
                      <Link to={step.link} className="small">
                        Go to Projects <i className="bi bi-arrow-right"></i>
                      </Link>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Language Tabs */}
      <div className="card">
        <div className="card-header">
          <ul className="nav nav-tabs card-header-tabs" role="tablist">
            {Object.entries(languages).map(([key, lang]) => (
              <li className="nav-item" role="presentation" key={key}>
                <button
                  className={`nav-link ${activeTab === key ? 'active' : ''}`}
                  onClick={() => setActiveTab(key)}
                  type="button"
                  role="tab"
                >
                  <i className={`${lang.icon} me-2`}></i>
                  {lang.label}
                </button>
              </li>
            ))}
          </ul>
        </div>
        <div className="card-body">
          {Object.entries(languages).map(([key, lang]) => (
            <div
              key={key}
              className={`tab-pane ${activeTab === key ? 'active' : 'd-none'}`}
              role="tabpanel"
            >
              {lang.samples.map((sample, idx) => (
                <div key={idx} className={idx > 0 ? 'mt-4' : ''}>
                  <div className="d-flex justify-content-between align-items-center mb-2">
                    <div>
                      <h5 className="mb-0">{sample.title}</h5>
                      <p className="text-muted small mb-0">{sample.description}</p>
                    </div>
                    <button
                      className="btn btn-sm btn-outline-secondary"
                      onClick={() => handleCopyCode(sample.code, `${key}-${idx}`)}
                    >
                      <i className={`bi bi-${copiedCode === `${key}-${idx}` ? 'check' : 'clipboard'} me-1`}></i>
                      {copiedCode === `${key}-${idx}` ? 'Copied!' : 'Copy'}
                    </button>
                  </div>
                  <pre className="bg-dark text-light p-3 rounded" style={{ fontSize: '0.85rem', maxHeight: '500px', overflow: 'auto' }}>
                    <code>{sample.code}</code>
                  </pre>
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Environment Variables Reference */}
      <div className="card mt-4">
        <div className="card-body">
          <h5 className="card-title mb-3">
            <i className="bi bi-gear me-2"></i>
            Environment Variables
          </h5>
          <table className="table table-sm">
            <thead>
              <tr>
                <th>Variable</th>
                <th>Description</th>
                <th>Example</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><code>LOGS_API_URL</code></td>
                <td>URL of the Logs service API</td>
                <td><code>http://localhost:8082</code></td>
              </tr>
              <tr>
                <td><code>LOGS_API_KEY</code></td>
                <td>API key from Projects page (Bearer token)</td>
                <td><code>proj_abc123...</code></td>
              </tr>
              <tr>
                <td><code>SERVICE_NAME</code></td>
                <td>Name of your service (appears in logs)</td>
                <td><code>my-api</code></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Performance Tips */}
      <div className="alert alert-info mt-4">
        <h6 className="alert-heading">
          <i className="bi bi-lightning-charge me-2"></i>
          Performance Tips
        </h6>
        <ul className="mb-0">
          <li><strong>Batch size:</strong> Use batches of 100-1000 logs for optimal performance (14,000-33,000 logs/second)</li>
          <li><strong>Flush interval:</strong> 5 seconds is recommended for balancing real-time visibility and throughput</li>
          <li><strong>Rate limits:</strong> 1000 requests per minute per API key (contact admin if you need higher limits)</li>
          <li><strong>Metadata:</strong> Keep metadata concise to reduce payload size and improve query performance</li>
        </ul>
      </div>
    </div>
  );
}
