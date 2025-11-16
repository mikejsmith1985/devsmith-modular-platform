let selectedMode = null;
let currentSessionId = null;
let currentCode = null;

// Mode selection and API trigger
document.querySelectorAll('.btn-select-mode').forEach(btn => {
  btn.addEventListener('click', async (e) => {
    e.preventDefault();
    
    selectedMode = e.target.dataset.mode;
    console.log(`Selected mode: ${selectedMode}`);
    
    // If we have code/session, trigger analysis immediately
    if (currentSessionId && currentCode) {
      await triggerReadingModeAnalysis(selectedMode);
    }
  });
});

/**
 * Trigger reading mode analysis via API
 */
async function triggerReadingModeAnalysis(mode) {
  const resultsContainer = document.getElementById('reading-mode-demo');
  if (!resultsContainer) return;
  
  // Show loading
  resultsContainer.innerHTML = `
    <div class="flex items-center gap-3 p-4">
      <span class="loading loading-spinner loading-sm"></span>
      <span>Analyzing code in ${mode} mode...</span>
    </div>
  `;
  
  try {
    let requestData = { code: currentCode };
    let endpoint = `/api/review/sessions/${currentSessionId}/modes/${mode}`;
    
    // Customize request based on mode
    if (mode === 'skim') {
      requestData = { repo_owner: 'devsmith', repo_name: 'platform' };
    } else if (mode === 'scan') {
      requestData = { query: 'error handling' };
    } else if (mode === 'detailed') {
      requestData = { file: 'main.go' };
    } else if (mode === 'critical') {
      requestData = { full_code: currentCode };
    }
    
    // Call API
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestData),
    });
    
    if (!response.ok) {
      throw new Error(`Analysis failed: ${response.statusText}`);
    }
    
    const result = await response.json();
    displayReadingModeResults(mode, result, resultsContainer);
    
  } catch (error) {
    resultsContainer.innerHTML = `
      <div class="alert alert-error">
        <span>Error: ${error.message}</span>
      </div>
    `;
  }
}

/**
 * Display reading mode results in the appropriate format
 */
function displayReadingModeResults(mode, result, container) {
  let html = '';
  
  if (mode === 'preview') {
    html = `
      <section class="card">
        <h3 class="text-xl font-bold mb-4">👁️ Preview Mode Results</h3>
        <div class="space-y-4">
          <div>
            <h4 class="font-semibold text-gray-700 dark:text-gray-300">Bounded Contexts</h4>
            <ul class="list-disc list-inside text-sm text-gray-600 dark:text-gray-400">
              ${result.BoundedContexts?.map(ctx => `<li>${ctx}</li>`).join('') || '<li>N/A</li>'}
            </ul>
          </div>
          <div>
            <h4 class="font-semibold text-gray-700 dark:text-gray-300">Tech Stack</h4>
            <div class="flex gap-2 flex-wrap">
              ${result.TechStack?.map(tech => `<span class="badge badge-primary">${tech}</span>`).join('') || 'N/A'}
            </div>
          </div>
          <div>
            <h4 class="font-semibold text-gray-700 dark:text-gray-300">Summary</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400">${result.Summary || 'N/A'}</p>
          </div>
        </div>
      </section>
    `;
  } else if (mode === 'skim') {
    html = `
      <section class="card">
        <h3 class="text-xl font-bold mb-4">⚡ Skim Mode Results</h3>
        <div class="space-y-4">
          <div>
            <h4 class="font-semibold">Functions</h4>
            <ul class="list-disc list-inside">
              ${result.Functions?.map(fn => `<li class="text-sm">${fn}</li>`).join('') || '<li>N/A</li>'}
            </ul>
          </div>
          <div>
            <h4 class="font-semibold">Key Imports</h4>
            <div class="flex gap-2 flex-wrap">
              ${result.Imports?.map(imp => `<code class="bg-gray-200 dark:bg-gray-700 px-2 py-1 rounded text-xs">${imp}</code>`).join('') || 'N/A'}
            </div>
          </div>
        </div>
      </section>
    `;
  } else if (mode === 'scan') {
    html = `
      <section class="card">
        <h3 class="text-xl font-bold mb-4">🔎 Scan Mode Results</h3>
        <div class="space-y-3">
          <p class="text-sm text-gray-600 dark:text-gray-400">Query: <strong>${result.Query}</strong></p>
          ${result.Matches?.map(match => `
            <div class="border-l-4 border-blue-500 pl-3">
              <p class="text-sm font-mono">${match.File}:${match.Line}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400">${match.Content}</p>
              <p class="text-xs text-blue-600">Relevance: ${(match.Relevance * 100).toFixed(0)}%</p>
            </div>
          `).join('') || '<p class="text-sm text-gray-500">No matches found</p>'}
        </div>
      </section>
    `;
  } else if (mode === 'detailed') {
    html = `
      <section class="card">
        <h3 class="text-xl font-bold mb-4">📖 Detailed Mode Results</h3>
        <div class="space-y-3">
          <p class="text-sm font-semibold">File: ${result.File}</p>
          ${result.LineByLine?.map(line => `
            <div class="bg-gray-50 dark:bg-gray-800 p-3 rounded">
              <p class="text-xs font-mono text-gray-700 dark:text-gray-300">Line ${line.LineNumber}: ${line.Code}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">${line.Explanation}</p>
            </div>
          `).join('') || '<p class="text-sm text-gray-500">No lines</p>'}
        </div>
      </section>
    `;
  } else if (mode === 'critical') {
    html = `
      <section class="card">
        <h3 class="text-xl font-bold mb-4">🔬 Critical Mode Results</h3>
        <div class="mb-4">
          <div class="text-2xl font-bold text-center">
            <span class="text-${result.OverallQuality >= 80 ? 'green' : result.OverallQuality >= 60 ? 'yellow' : 'red'}-600">
              ${result.OverallQuality}%
            </span>
          </div>
          <p class="text-xs text-center text-gray-500">${result.Summary}</p>
        </div>
        <div class="space-y-2">
          ${result.Issues?.map(issue => `
            <div class="alert alert-${issue.Severity === 'CRITICAL' ? 'error' : 'warning'}">
              <div>
                <p class="font-semibold text-sm">${issue.Category}: ${issue.Description}</p>
                <p class="text-xs opacity-80">Line ${issue.Line}</p>
                <p class="text-xs opacity-80">Suggestion: ${issue.Suggestion}</p>
              </div>
            </div>
          `).join('') || '<p class="text-sm text-green-600">No issues found!</p>'}
        </div>
      </section>
    `;
  }
  
  container.innerHTML = html;
}

// Form submission for session creation
document.getElementById('review-session-form')?.addEventListener('submit', async (e) => {
  e.preventDefault();

  const formData = new FormData(e.target);
  const pastedCode = formData.get('pasted_code');
  const githubUrl = formData.get('github_url');
  const uploadedFile = formData.get('file');

  // Validate that at least one input method is provided
  if (!pastedCode && !githubUrl && !uploadedFile) {
    showError('Please provide code via paste, GitHub URL, or file upload');
    return;
  }

  // Clear any previous errors
  clearError();

  // Show progress indicator
  showLoading('session');

  try {
    // Prepare request data
    const requestData = {
      code_source: 'paste', // Default to paste
      pasted_code: pastedCode || '',
      title: 'Code Review Session'
    };

    if (githubUrl) {
      requestData.code_source = 'github';
      requestData.github_url = githubUrl;
    } else if (uploadedFile) {
      requestData.code_source = 'upload';
      // File upload would need different handling
    }

    // Create session
    const response = await fetch('/review/sessions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData),
    });

    if (!response.ok) {
      throw new Error(`Session creation failed: ${response.statusText}`);
    }

    const result = await response.json();
    currentSessionId = result.session_id;
    currentCode = pastedCode || ''; // Assuming pastedCode is the primary source for now

    // Update progress to show session created
    updateProgress(10);
    document.querySelector('.progress-message').textContent = 'Session created, awaiting analysis updates...';

    // If a simulated interval exists, stop it (showLoading may have started one)
    if (window.progressInterval) {
      clearInterval(window.progressInterval);
      window.progressInterval = null;
    }

    // Connect to SSE endpoint to receive live progress updates for this session
    const sessionId = result.session_id;
    const sseUrl = `/review/sessions/${sessionId}/progress`;
    const es = new EventSource(sseUrl);

    es.addEventListener('progress', (evt) => {
      try {
        const data = JSON.parse(evt.data);
        if (typeof data.percent === 'number') {
          updateProgress(data.percent);
        }
        if (data.message) {
          const msgEl = document.querySelector('.progress-message');
          if (msgEl) msgEl.textContent = data.message;
        }
        if (data.percent >= 100) {
          // Completed - close connection and finalize
          es.close();
          setTimeout(() => {
            hideLoading();
            alert(`Analysis complete. Session ID: ${sessionId}`);
            // Optionally redirect to results page
            // window.location.href = `/analysis?session=${sessionId}`;
          }, 400);
        }
      } catch (err) {
        console.error('Failed to parse SSE progress data', err);
      }
    });

    es.onerror = (err) => {
      console.warn('SSE connection error', err);
      // On error, fallback to finishing the progress and notifying the user
      es.close();
      hideLoading();
      alert(`Session created (ID: ${sessionId}), but live updates failed.`);
    };

  } catch (error) {
    hideLoading();
    showError(`Error: ${error.message}`);
  }
});

function showError(message) {
  const errorDiv = document.querySelector('.error-message');
  if (errorDiv) {
    errorDiv.textContent = message;
    errorDiv.classList.remove('hidden');
  }
}

function clearError() {
  const errorDiv = document.querySelector('.error-message');
  if (errorDiv) {
    errorDiv.textContent = '';
    errorDiv.classList.add('hidden');
  }
}


// Show dynamic progress indicator in the main UI (not overlay)
function showLoading(mode) {
  // Remove any existing indicator
  const container = document.getElementById('progress-indicator-container');
  if (!container) return;
  container.innerHTML = '';

  // Create progress indicator element (Templ markup style)
  const wrapper = document.createElement('div');
  wrapper.className = 'progress-indicator flex items-center gap-3 my-4';
  wrapper.setAttribute('role', 'status');
  wrapper.setAttribute('aria-live', 'polite');

  const spinner = document.createElement('span');
  spinner.className = 'loading loading-spinner loading-md text-primary';
  spinner.setAttribute('aria-label', 'Loading');
  wrapper.appendChild(spinner);

  const msg = document.createElement('span');
  msg.className = 'progress-message font-medium';
  msg.textContent = `Analyzing code in ${capitalizeMode(mode)} mode...`;
  wrapper.appendChild(msg);

  const progress = document.createElement('progress');
  progress.className = 'progress progress-primary w-40';
  progress.value = 0;
  progress.max = 100;
  progress.id = 'progress-bar';
  wrapper.appendChild(progress);

  const percent = document.createElement('span');
  percent.className = 'progress-percent text-xs text-gray-500';
  percent.id = 'progress-percent';
  percent.textContent = '0%';
  wrapper.appendChild(percent);

  container.appendChild(wrapper);

  // Simulate progress for demo (replace with real updates via polling/websocket)
  let current = 0;
  window.progressInterval = setInterval(() => {
    if (current < 95) {
      current += Math.floor(Math.random() * 5) + 1;
      if (current > 95) current = 95;
      updateProgress(current);
    }
  }, 400);
}

function updateProgress(percentValue) {
  const bar = document.getElementById('progress-bar');
  const percent = document.getElementById('progress-percent');
  if (bar && percent) {
    bar.value = percentValue;
    percent.textContent = `${percentValue}%`;
  }
}

function hideLoading() {
  // Remove progress indicator and clear interval
  const container = document.getElementById('progress-indicator-container');
  if (container) container.innerHTML = '';
  if (window.progressInterval) {
    clearInterval(window.progressInterval);
    window.progressInterval = null;
  }
}

function capitalizeMode(mode) {
  return mode.charAt(0).toUpperCase() + mode.slice(1);
}
