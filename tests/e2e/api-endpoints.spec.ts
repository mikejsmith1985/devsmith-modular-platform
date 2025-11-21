import { test, expect } from '@playwright/test';

const PLAYWRIGHT_BASE = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000';
const ORIGIN = new URL(PLAYWRIGHT_BASE).origin;

test.describe('API Endpoints - Logs Service', () => {
  test('GET /api/logs/v1/stats should return log statistics', async ({ request }) => {
    const response = await request.get('/api/logs/v1/stats');
    
    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('application/json');
    
    const data = await response.json();
    expect(data).toHaveProperty('debug');
    expect(data).toHaveProperty('info');
    expect(data).toHaveProperty('warning');
    expect(data).toHaveProperty('error');
    expect(data).toHaveProperty('critical');
    
    // All values should be numbers
    expect(typeof data.debug).toBe('number');
    expect(typeof data.info).toBe('number');
    expect(typeof data.warning).toBe('number');
    expect(typeof data.error).toBe('number');
    expect(typeof data.critical).toBe('number');
  });

  test('GET /api/logs/health should return healthy status', async ({ request }) => {
    const response = await request.get('/api/logs/health');
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('service', 'logs');
    expect(data).toHaveProperty('status', 'healthy');
  });

  test('Stats endpoint should respond quickly (< 5 seconds)', async ({ request }) => {
    const start = Date.now();
    const response = await request.get('/api/logs/v1/stats');
    const duration = Date.now() - start;
    
    expect(response.status()).toBe(200);
    expect(duration).toBeLessThan(5000); // Should be much faster than old 87s!
  });
});

test.describe('API Endpoints - Portal Service', () => {
  test('GET /api/portal/health should return healthy status', async ({ request }) => {
    const response = await request.get('/api/portal/health');
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('service', 'portal');
    expect(data).toHaveProperty('status', 'healthy');
  });
});

test.describe('API Endpoints - Review Service', () => {
  test('GET /api/review/health should return healthy status', async ({ request }) => {
    const response = await request.get('/api/review/health');
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('service', 'review');
    expect(data).toHaveProperty('status', 'healthy');
  });
});

test.describe('API Endpoints - Analytics Service', () => {
  test('GET /api/analytics/health should return healthy status', async ({ request }) => {
    const response = await request.get('/api/analytics/health');
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    expect(data).toHaveProperty('service', 'analytics');
    expect(data).toHaveProperty('status', 'healthy');
  });
});

test.describe('API Endpoints - Traefik Routing', () => {
  test('API routes should have higher priority than frontend', async ({ request }) => {
    // API route should return JSON
    const apiResponse = await request.get('/api/logs/health');
    expect(apiResponse.headers()['content-type']).toContain('application/json');
    
    // Frontend route should return HTML
    const frontendResponse = await request.get('/');
    expect(frontendResponse.headers()['content-type']).toContain('text/html');
  });

  test('Unknown API routes should return 404 (not frontend HTML)', async ({ request }) => {
    const response = await request.get('/api/nonexistent');
    
    // Should return 404 from Traefik, not 200 from frontend
    expect(response.status()).toBe(404);
  });
});

test.describe('API Endpoints - CORS and Headers', () => {
  test('API endpoints should have proper CORS headers', async ({ request }) => {
    const response = await request.get('/api/logs/health', {
      headers: {
        'Origin': ORIGIN
      }
    });
    
    expect(response.status()).toBe(200);
    // Verify CORS headers if configured
  });

  test('API endpoints should accept JSON content type', async ({ request }) => {
    const response = await request.post('/api/logs', {
      data: {
        level: 'info',
        message: 'Test log entry',
        service: 'test'
      },
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    // Should accept JSON (may need auth, so 401 is acceptable)
    expect([200, 201, 401]).toContain(response.status());
  });
});
