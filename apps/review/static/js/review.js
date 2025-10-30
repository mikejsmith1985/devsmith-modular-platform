let selectedMode = null;

// Mode selection (keeping for future use)
document.querySelectorAll('.btn-select-mode').forEach(btn => {
  btn.addEventListener('click', (e) => {
    selectedMode = e.target.dataset.mode;
    // Mode selection logic can be added here later
  });
});

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
