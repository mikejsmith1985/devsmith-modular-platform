import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

export default function OAuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [error, setError] = useState(null);

  console.log('[OAuthCallback] Component mounted');
  console.log('[OAuthCallback] Search params:', searchParams.toString());

  useEffect(() => {
    console.log('[OAuthCallback] useEffect triggered');
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const errorParam = searchParams.get('error');

    console.log('[OAuthCallback] Code from URL:', code);
    console.log('[OAuthCallback] State from URL:', state);
    console.log('[OAuthCallback] Error from URL:', errorParam);

    if (errorParam) {
      console.log('[OAuthCallback] Error detected, showing error and redirecting');
      setError('GitHub authentication failed. Please try again.');
      setTimeout(() => navigate('/login'), 3000);
      return;
    }

    if (!code) {
      console.log('[OAuthCallback] No authorization code detected, showing error and redirecting');
      setError('No authorization code received.');
      setTimeout(() => navigate('/login'), 3000);
      return;
    }

    // Validate state (CSRF protection)
    const storedState = sessionStorage.getItem('oauth_state');
    if (!state || state !== storedState) {
      console.error('[PKCE] State mismatch - possible CSRF attack');
      console.error('[PKCE] Expected:', storedState, 'Received:', state);
      setError('Security validation failed. Please try again.');
      sessionStorage.clear(); // Clear PKCE data
      setTimeout(() => navigate('/login'), 3000);
      return;
    }

    // Get PKCE verifier
    const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
    if (!codeVerifier) {
      console.error('[PKCE] Missing code verifier');
      setError('Security validation failed. Please try again.');
      setTimeout(() => navigate('/login'), 3000);
      return;
    }

    // Exchange code for token (send verifier to backend)
    const exchangeCodeForToken = async () => {
      try {
        console.log('[PKCE] Exchanging code for token...');
        const response = await fetch('/api/portal/auth/token', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            code,
            state,
            code_verifier: codeVerifier,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
          console.error('[PKCE] Token exchange failed:', errorData);
          throw new Error(errorData.error || 'Token exchange failed');
        }

        const data = await response.json();
        console.log('[PKCE] Token exchange successful');
        
        // Store JWT
        localStorage.setItem('devsmith_token', data.token);
        console.log('[PKCE] Token stored in localStorage');
        
        // Clear PKCE data
        sessionStorage.removeItem('pkce_code_verifier');
        sessionStorage.removeItem('oauth_state');
        console.log('[PKCE] PKCE data cleared from sessionStorage');
        
        // Redirect to dashboard
        console.log('[PKCE] Redirecting to dashboard');
        navigate('/');
      } catch (err) {
        console.error('[PKCE] Token exchange error:', err);
        setError(`Failed to complete authentication: ${err.message}`);
        sessionStorage.clear();
        setTimeout(() => navigate('/login'), 3000);
      }
    };

    exchangeCodeForToken();
  }, [searchParams, navigate]);

  return (
    <div className="container">
      <div className="row justify-content-center align-items-center" style={{ minHeight: '100vh' }}>
        <div className="col-md-6 col-lg-4 text-center">
          {error ? (
            <div>
              <i className="bi bi-exclamation-triangle text-danger" style={{ fontSize: '3rem' }}></i>
              <h3 className="mt-3">Authentication Error</h3>
              <p className="text-muted">{error}</p>
              <p className="text-muted">Redirecting to login...</p>
            </div>
          ) : (
            <div>
              <div className="spinner-border text-primary" role="status" style={{ width: '3rem', height: '3rem' }}>
                <span className="visually-hidden">Loading...</span>
              </div>
              <h3 className="mt-3">Authenticating with GitHub...</h3>
              <p className="text-muted">Please wait while we complete your login.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
