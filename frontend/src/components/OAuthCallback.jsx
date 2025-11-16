import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { decryptVerifier } from '../utils/pkce';

export default function OAuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [error, setError] = useState(null);

  console.log('[OAuthCallback] Component mounted');
  console.log('[OAuthCallback] Search params:', searchParams.toString());

  useEffect(() => {
    console.log('[OAuthCallback] useEffect triggered');
    const code = searchParams.get('code');
    const encryptedState = searchParams.get('state');
    const errorParam = searchParams.get('error');

    console.log('[OAuthCallback] Code from URL:', code);
    console.log('[OAuthCallback] Encrypted state from URL:', encryptedState);
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

    if (!encryptedState) {
      console.error('[PKCE] Missing encrypted state parameter');
      setError('Security validation failed. Please try again.');
      setTimeout(() => navigate('/login'), 3000);
      return;
    }

    // Exchange code for token (decrypt verifier from state)
    const exchangeCodeForToken = async () => {
      try {
        console.log('[PKCE] Decrypting verifier from state...');
        
        // Decrypt verifier from state (validates timestamp automatically)
        const codeVerifier = await decryptVerifier(encryptedState);
        console.log('[PKCE] Verifier decrypted successfully');
        
        console.log('[PKCE] Exchanging code for token...');
        const response = await fetch('/api/portal/auth/token', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            code,
            state: encryptedState, // Send for audit logging
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
        
        // Redirect to dashboard with page reload to trigger AuthContext
        console.log('[PKCE] Redirecting to dashboard');
        window.location.href = '/';
      } catch (err) {
        console.error('[PKCE] Token exchange error:', err);
        
        // Provide user-friendly error messages
        let errorMessage = 'Failed to complete authentication';
        if (err.message.includes('expired')) {
          errorMessage = 'Login session expired (>10 minutes). Please try again.';
        } else if (err.message.includes('Invalid or tampered')) {
          errorMessage = 'Security validation failed. Please try again.';
        } else {
          errorMessage = `${errorMessage}: ${err.message}`;
        }
        
        setError(errorMessage);
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
