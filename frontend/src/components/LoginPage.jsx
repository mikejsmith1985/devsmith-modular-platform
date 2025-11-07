import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { generateCodeVerifier, generateCodeChallenge, generateState } from '../utils/pkce';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [pkceError, setPkceError] = useState(null);
  const { login, isAuthenticated, error } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/');
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    await login(email, password);
  };

  const handleGitHubLogin = async () => {
    try {
      // Generate PKCE parameters
      const codeVerifier = generateCodeVerifier();
      const codeChallenge = await generateCodeChallenge(codeVerifier);
      const state = generateState();

      // Store verifier and state in sessionStorage (temporary, cleared on tab close)
      sessionStorage.setItem('pkce_code_verifier', codeVerifier);
      sessionStorage.setItem('oauth_state', state);

      // Get GitHub client ID from environment
      const clientId = import.meta.env.VITE_GITHUB_CLIENT_ID;
      if (!clientId) {
        throw new Error('GitHub Client ID not configured');
      }

      // Build GitHub OAuth URL with PKCE
      const params = new URLSearchParams({
        client_id: clientId,
        redirect_uri: window.location.origin + '/auth/callback',
        scope: 'user:email read:user',
        state: state,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256',
      });

      const authURL = `https://github.com/login/oauth/authorize?${params}`;
      console.log('[PKCE] Redirecting to GitHub with code_challenge');
      window.location.href = authURL;
    } catch (error) {
      console.error('[PKCE] Failed to generate PKCE parameters:', error);
      setPkceError('Failed to initiate login. Please try again.');
    }
  };

  return (
    <div className="container">
      <div className="row justify-content-center align-items-center" style={{ minHeight: '100vh' }}>
        <div className="col-md-6 col-lg-4">
          <div className="card shadow">
            <div className="card-body p-5">
              <h2 className="text-center mb-4">
                <i className="bi bi-code-square text-primary" style={{ fontSize: '2.5rem' }}></i>
                <div className="mt-2">DevSmith Platform</div>
              </h2>

              {(error || pkceError) && (
                <div className="alert alert-danger" role="alert">
                  {error || pkceError}
                </div>
              )}

              <form onSubmit={handleSubmit}>
                <div className="mb-3">
                  <label htmlFor="email" className="form-label">Email</label>
                  <input
                    type="email"
                    className="form-control"
                    id="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                  />
                </div>

                <div className="mb-3">
                  <label htmlFor="password" className="form-label">Password</label>
                  <input
                    type="password"
                    className="form-control"
                    id="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                  />
                </div>

                <button type="submit" className="btn btn-primary w-100 mb-3">
                  Login
                </button>
              </form>

              <div className="text-center">
                <div className="mb-2">or</div>
                <button
                  type="button"
                  className="btn btn-dark w-100"
                  onClick={handleGitHubLogin}
                >
                  <i className="bi bi-github me-2"></i>
                  Login with GitHub
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
