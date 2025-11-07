// API utilities for DevSmith frontend
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000';

class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

// Generic API fetch with error handling
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  
  const defaultOptions = {
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for session auth
  };

  const response = await fetch(url, { ...defaultOptions, ...options });
  
  if (!response.ok) {
    const errorText = await response.text();
    throw new ApiError(`HTTP ${response.status}: ${errorText}`, response.status);
  }

  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return response.json();
  }
  return response.text();
}

// Review API endpoints
export const reviewApi = {
  // Get available models
  getModels: () => apiRequest('/api/review/models'),
  
  // Create new review session
  createSession: (data) => apiRequest('/api/review/sessions', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  
  // Run analysis in different modes
  runPreview: (sessionId, code, model) => apiRequest('/api/review/modes/preview', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, code, model }),
  }),
  
  runSkim: (sessionId, code, model) => apiRequest('/api/review/modes/skim', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, code, model }),
  }),
  
  runScan: (sessionId, code, model, query) => apiRequest('/api/review/modes/scan', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, code, model, query }),
  }),
  
  runDetailed: (sessionId, code, model) => apiRequest('/api/review/modes/detailed', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, code, model }),
  }),
  
  runCritical: (sessionId, code, model) => apiRequest('/api/review/modes/critical', {
    method: 'POST',
    body: JSON.stringify({ session_id: sessionId, code, model }),
  }),
};

// Logs API endpoints
export const logsApi = {
  getLogs: (params) => {
    const queryString = new URLSearchParams(params).toString();
    return apiRequest(`/api/logs${queryString ? '?' + queryString : ''}`);
  },
  
  getStats: () => apiRequest('/api/logs/stats'),
};

// Analytics API endpoints
export const analyticsApi = {
  getTrends: () => apiRequest('/api/analytics/trends'),
  getTopIssues: () => apiRequest('/api/analytics/top-issues'),
  getAnomalies: () => apiRequest('/api/analytics/anomalies'),
};

export { ApiError };