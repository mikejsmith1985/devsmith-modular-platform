// Logout functionality
document.getElementById('logout-btn').addEventListener('click', async () => {
  // Clear JWT from localStorage
  localStorage.removeItem('devsmith_token');

  // Redirect to login
  window.location.href = '/login';
});

// Optional: Check service health on load
async function checkServiceHealth() {
  const services = [
    { name: 'Review', url: 'http://localhost:8081/health' },
    { name: 'Logs', url: 'http://localhost:8082/health' },
    { name: 'Analytics', url: 'http://localhost:8083/health' },
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