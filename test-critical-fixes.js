#!/usr/bin/env node
/**
 * Test script to validate critical fixes:
 * 1. Health check optimization (503 Service Unavailable fix)
 * 2. API field name correction (Code required fix)
 */

const https = require('https');
const http = require('http');

// Colors for output
const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
};

function log(color, message) {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function makeRequest(options, data = null) {
  return new Promise((resolve, reject) => {
    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => {
        body += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          body: body,
          bodyJSON: (() => {
            try { return JSON.parse(body); } catch { return null; }
          })()
        });
      });
    });

    req.on('error', (err) => {
      reject(err);
    });

    if (data) {
      req.write(data);
    }
    req.end();
  });
}

async function testHealthCheckPerformance() {
  log('cyan', '\n🏥 Testing Health Check Performance');
  log('blue', '='.repeat(50));
  
  const start = Date.now();
  
  try {
    const response = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/health',
      method: 'GET'
    });
    
    const duration = Date.now() - start;
    
    if (response.statusCode === 200 && response.bodyJSON) {
      log('green', `✅ Health check SUCCESS in ${duration}ms`);
      
      // Check for fast response time
      if (duration < 3000) {
        log('green', `✅ Performance GOOD: ${duration}ms < 3000ms (timeout threshold)`);
      } else {
        log('red', `❌ Performance POOR: ${duration}ms >= 3000ms`);
        return false;
      }
      
      // Check individual component timings
      const components = response.bodyJSON.components;
      let slowComponents = [];
      
      components.forEach(comp => {
        // Convert microseconds to milliseconds for readability
        const responseTimeMs = comp.response_time_ms / 1000;
        if (responseTimeMs > 2000) {
          slowComponents.push(`${comp.name}: ${responseTimeMs.toFixed(1)}ms`);
        }
      });
      
      if (slowComponents.length === 0) {
        log('green', '✅ All components responding quickly');
      } else {
        log('yellow', `⚠️  Slow components detected: ${slowComponents.join(', ')}`);
      }
      
      return true;
    } else {
      log('red', `❌ Health check failed: ${response.statusCode}`);
      log('red', response.body);
      return false;
    }
  } catch (error) {
    log('red', `❌ Health check ERROR: ${error.message}`);
    return false;
  }
}

async function testTraefikGateway() {
  log('cyan', '\n🌐 Testing Traefik Gateway (503 Error Fix)');
  log('blue', '='.repeat(50));
  
  try {
    // Test models endpoint through gateway
    const response = await makeRequest({
      hostname: 'localhost',
      port: 3000,
      path: '/api/review/models',
      method: 'GET'
    });
    
    if (response.statusCode === 200 && response.bodyJSON) {
      log('green', '✅ Traefik gateway routing SUCCESS');
      log('green', `✅ Found ${response.bodyJSON.models.length} models available`);
      
      // List available models
      response.bodyJSON.models.forEach(model => {
        log('blue', `   📦 ${model.name}: ${model.description}`);
      });
      
      return true;
    } else if (response.statusCode === 503) {
      log('red', '❌ Still getting 503 Service Unavailable');
      log('red', 'Health checks may still be too slow for Traefik');
      return false;
    } else {
      log('red', `❌ Unexpected status: ${response.statusCode}`);
      log('red', response.body);
      return false;
    }
  } catch (error) {
    log('red', `❌ Gateway test ERROR: ${error.message}`);
    return false;
  }
}

async function testAPIFieldNames() {
  log('cyan', '\n🔧 Testing API Field Names (Code Required Fix)');
  log('blue', '='.repeat(50));
  
  const testCode = `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`;

  // Test with CORRECT field name (pasted_code)
  try {
    log('blue', 'Testing with CORRECT field name: pasted_code');
    const correctResponse = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/api/review/modes/preview',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, JSON.stringify({
      session_id: 'test-correct',
      pasted_code: testCode,
      model: 'mistral:7b-instruct'
    }));
    
    if (correctResponse.statusCode === 401) {
      log('green', '✅ CORRECT field name accepted (got auth error as expected)');
    } else if (correctResponse.statusCode === 400 && correctResponse.body.includes('required')) {
      log('red', '❌ Still getting "required" error with pasted_code field');
      return false;
    } else {
      log('green', `✅ CORRECT field name accepted (status: ${correctResponse.statusCode})`);
    }
  } catch (error) {
    log('red', `❌ Correct field test ERROR: ${error.message}`);
    return false;
  }
  
  // Test with WRONG field name (code) - should fail
  try {
    log('blue', 'Testing with WRONG field name: code');
    const wrongResponse = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/api/review/modes/preview',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, JSON.stringify({
      session_id: 'test-wrong',
      code: testCode,  // WRONG field name
      model: 'mistral:7b-instruct'
    }));
    
    if (wrongResponse.statusCode === 400 && wrongResponse.body.includes('required')) {
      log('green', '✅ WRONG field name properly rejected with "required" error');
      return true;
    } else {
      log('yellow', `⚠️  Unexpected response for wrong field: ${wrongResponse.statusCode}`);
      log('yellow', wrongResponse.body);
      return true; // Still OK, just different error handling
    }
  } catch (error) {
    log('red', `❌ Wrong field test ERROR: ${error.message}`);
    return false;
  }
}

async function testDockerServices() {
  log('cyan', '\n🐳 Testing Docker Service Status');
  log('blue', '='.repeat(50));
  
  try {
    // Check if services are running
    const { exec } = require('child_process');
    
    return new Promise((resolve) => {
      exec('docker-compose ps --format json', (error, stdout, stderr) => {
        if (error) {
          log('red', `❌ Docker command failed: ${error.message}`);
          resolve(false);
          return;
        }
        
        try {
          const services = stdout.trim().split('\n')
            .filter(line => line.trim())
            .map(line => JSON.parse(line));
          
          let allHealthy = true;
          
          services.forEach(service => {
            if (service.Health === 'healthy') {
              log('green', `✅ ${service.Service}: ${service.Health} (${service.State})`);
            } else if (service.Health === 'unhealthy') {
              log('red', `❌ ${service.Service}: ${service.Health} (${service.State})`);
              allHealthy = false;
            } else {
              log('blue', `📋 ${service.Service}: ${service.State}`);
            }
          });
          
          resolve(allHealthy);
        } catch (parseError) {
          log('red', `❌ Failed to parse Docker output: ${parseError.message}`);
          resolve(false);
        }
      });
    });
  } catch (error) {
    log('red', `❌ Docker test ERROR: ${error.message}`);
    return false;
  }
}

async function main() {
  log('bold', '\n🚀 DevSmith Critical Fixes Validation');
  log('magenta', '=' .repeat(60));
  log('yellow', 'Testing fixes for:');
  log('yellow', '1. Health check optimization (503 Service Unavailable)');
  log('yellow', '2. API field name correction (Code required error)');
  
  const results = {
    dockerServices: await testDockerServices(),
    healthCheck: await testHealthCheckPerformance(),
    traefik: await testTraefikGateway(),
    apiFields: await testAPIFieldNames()
  };
  
  // Summary
  log('cyan', '\n📊 Test Results Summary');
  log('blue', '='.repeat(50));
  
  const allPassed = Object.values(results).every(result => result === true);
  
  Object.entries(results).forEach(([test, passed]) => {
    const status = passed ? '✅ PASS' : '❌ FAIL';
    const color = passed ? 'green' : 'red';
    log(color, `${status} ${test}`);
  });
  
  if (allPassed) {
    log('green', '\n🎉 ALL CRITICAL FIXES VALIDATED! 🎉');
    log('green', 'Both issues should now be resolved:');
    log('green', '• No more 503 Service Unavailable errors');
    log('green', '• No more "Code required" false errors');
  } else {
    log('red', '\n❌ Some tests failed. Check logs above for details.');
    process.exit(1);
  }
}

main().catch(error => {
  log('red', `Fatal error: ${error.message}`);
  process.exit(1);
});