// Copy analysis to clipboard
document.getElementById('copy-btn')?.addEventListener('click', async () => {
  const content = document.getElementById('analysis-content').innerText;

  try {
    await navigator.clipboard.writeText(content);

    // Visual feedback
    const btn = document.getElementById('copy-btn');
    const originalText = btn.textContent;
    btn.textContent = '✅';
    btn.classList.add('copied');

    setTimeout(() => {
      btn.textContent = originalText;
      btn.classList.remove('copied');
    }, 2000);

  } catch (err) {
    alert('Failed to copy to clipboard');
  }
});
