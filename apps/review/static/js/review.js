let selectedMode = null;

// Mode selection
document.querySelectorAll('.btn-select-mode').forEach(btn => {
  btn.addEventListener('click', (e) => {
    selectedMode = e.target.dataset.mode;
    document.getElementById('selected-mode').value = selectedMode;

    // Hide mode selector, show repo input
    document.querySelector('.mode-selector').classList.add('hidden');
    document.getElementById('repo-input-section').classList.remove('hidden');

    // Update form title with selected mode
    document.querySelector('#repo-input-section h2').textContent =
      `Repository Details (${capitalizeMode(selectedMode)} Mode)`;
  });
});

// Back button
document.getElementById('back-btn')?.addEventListener('click', () => {
  document.querySelector('.mode-selector').classList.remove('hidden');
  document.getElementById('repo-input-section').classList.add('hidden');
  selectedMode = null;
  document.getElementById('review-form').reset();
});

// Form submission
document.getElementById('review-form')?.addEventListener('submit', async (e) => {
  e.preventDefault();

  const repoUrl = document.getElementById('repository-url').value;
  const branch = document.getElementById('branch').value || 'main';
  const commitSha = document.getElementById('commit-sha').value || null;

  // Validate GitHub URL
  if (!repoUrl.startsWith('https://github.com/')) {
    alert('Please enter a valid GitHub repository URL');
    return;
  }

  if (!selectedMode) {
    alert('Please select a reading mode');
    return;
  }

  // Show loading overlay
  showLoading(selectedMode);

  try {
    // Call appropriate API endpoint based on selected mode
    const response = await fetch(`/api/v1/review/${selectedMode}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        repository_url: repoUrl,
        branch: branch,
        commit_sha: commitSha,
      }),
    });

    if (!response.ok) {
      throw new Error(`Analysis failed: ${response.statusText}`);
    }

    const result = await response.json();

    // Redirect to analysis results page with data
    const params = new URLSearchParams({
      mode: result.mode,
      repo: result.repository,
      branch: result.branch,
      analysis: result.analysis,
    });

    window.location.href = `/analysis?${params.toString()}`;

  } catch (error) {
    hideLoading();
    alert(`Error: ${error.message}`);
  }
});

function showLoading(mode) {
  const overlay = document.createElement('div');
  overlay.id = 'loading-overlay';
  overlay.className = 'loading-overlay';
  overlay.innerHTML = `
    <div class="loading-content">
      <div class="spinner"></div>
      <h2>Analyzing Repository...</h2>
      <p>Running <strong>${capitalizeMode(mode)}</strong> mode analysis</p>
      <p class="loading-message">This may take 30 seconds to 2 minutes.</p>
    </div>
  `;
  document.body.appendChild(overlay);
}

function hideLoading() {
  const overlay = document.getElementById('loading-overlay');
  if (overlay) overlay.remove();
}

function capitalizeMode(mode) {
  return mode.charAt(0).toUpperCase() + mode.slice(1);
}
