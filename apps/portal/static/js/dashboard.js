// Logout functionality
document.getElementById('logout-btn').addEventListener('click', async () => {
  // Clear JWT from localStorage
  localStorage.removeItem('devsmith_token');

  // Redirect to login
  window.location.href = '/login';
});

// Optional: Check service health on load
async function checkServiceHealth() {
  // Use gateway URL dynamically (works in Docker and production)
  const baseURL = window.location.origin; // Gets gateway URL automatically
  const services = [
    { name: 'Review', url: `${baseURL}/api/review/health` },
    { name: 'Logs', url: `${baseURL}/api/logs/health` },
    { name: 'Analytics', url: `${baseURL}/api/analytics/health` },
  ];

  for (const service of services) {
    try {
      const response = await fetch(service.url);
      if (response.ok) {
        console.log(`${service.name} service is healthy`);
      }
    } catch (err) {
      console.warn(`${service.name} service is not responding`);
    }
  }
}

// Run health check on page load
checkServiceHealth();