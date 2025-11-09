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
export async function apiRequest(endpoint, options = {}) {
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
  runPreview: (sessionId, code, model, userMode = 'intermediate', outputMode = 'quick') => apiRequest('/api/review/modes/preview', {
    method: 'POST',
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
  }),
  
  runSkim: (sessionId, code, model, userMode = 'intermediate', outputMode = 'quick') => apiRequest('/api/review/modes/skim', {
    method: 'POST',
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
  }),
  
  runScan: (sessionId, code, model, query, userMode = 'intermediate', outputMode = 'quick') => apiRequest('/api/review/modes/scan', {
    method: 'POST',
    body: JSON.stringify({ pasted_code: code, model, query, user_mode: userMode, output_mode: outputMode }),
  }),
  
  runDetailed: (sessionId, code, model, userMode = 'intermediate', outputMode = 'quick') => apiRequest('/api/review/modes/detailed', {
    method: 'POST',
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
  }),
  
  runCritical: (sessionId, code, model, userMode = 'intermediate', outputMode = 'quick') => apiRequest('/api/review/modes/critical', {
    method: 'POST',
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
  }),

  // GitHub Integration API endpoints (Phase 1)
  // Fetch repository tree structure
  githubGetTree: (url, branch = 'main') => {
    const params = new URLSearchParams({ url, branch });
    return apiRequest(`/api/review/github/tree?${params.toString()}`);
  },
  
  // Fetch individual file content
  githubGetFile: (url, path, branch = 'main') => {
    const params = new URLSearchParams({ url, path, branch });
    return apiRequest(`/api/review/github/file?${params.toString()}`);
  },
  
  // Quick repo scan (fetches 5-8 core files)
  githubQuickScan: (url, branch = 'main') => {
    const params = new URLSearchParams({ url, branch });
    return apiRequest(`/api/review/github/quick-scan?${params.toString()}`);
  },

  // Prompt Management API endpoints (Phase 4)
  // Get effective prompt (user custom or system default)
  getPrompt: (mode, userLevel = 'intermediate', outputMode = 'quick') => {
    const params = new URLSearchParams({ mode, user_level: userLevel, output_mode: outputMode });
    return apiRequest(`/api/review/prompts?${params.toString()}`);
  },
  
  // Save custom prompt
  savePrompt: (data) => apiRequest('/api/review/prompts', {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
  
  // Factory reset to system default
  resetPrompt: (mode, userLevel = 'intermediate', outputMode = 'quick') => {
    const params = new URLSearchParams({ mode, user_level: userLevel, output_mode: outputMode });
    return apiRequest(`/api/review/prompts?${params.toString()}`, {
      method: 'DELETE',
    });
  },
  
  // Get prompt execution history
  getPromptHistory: (limit = 50) => {
    const params = new URLSearchParams({ limit: limit.toString() });
    return apiRequest(`/api/review/prompts/history?${params.toString()}`);
  },
  
  // Rate prompt execution
  rateExecution: (executionId, rating) => apiRequest(`/api/review/prompts/${executionId}/rate`, {
    method: 'POST',
    body: JSON.stringify({ rating }),
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