// Utility to fetch installed Ollama models from backend

import { apiRequest } from './api';

export async function fetchOllamaModels() {
  try {
    const data = await apiRequest('/api/portal/llm-configs/ollama-models');
    return Array.isArray(data.models) ? data.models : [];
  } catch (err) {
    return [];
  }
}
