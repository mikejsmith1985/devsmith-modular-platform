/**
 * PKCE helper functions for OAuth 2.0 Authorization Code Flow
 * Implements RFC 7636 - Proof Key for Code Exchange
 */

/**
 * Generate a cryptographically random code verifier (43-128 chars)
 * @returns {string} Base64-URL encoded code verifier
 */
export function generateCodeVerifier() {
  const array = new Uint8Array(32);
  window.crypto.getRandomValues(array);
  return base64URLEncode(array);
}

/**
 * Generate SHA-256 code challenge from verifier
 * @param {string} verifier - The code verifier
 * @returns {Promise<string>} Base64-URL encoded code challenge
 */
export async function generateCodeChallenge(verifier) {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await window.crypto.subtle.digest('SHA-256', data);
  return base64URLEncode(new Uint8Array(digest));
}

/**
 * Generate random state for CSRF protection
 * @returns {string} Base64-URL encoded state parameter
 */
export function generateState() {
  const array = new Uint8Array(16);
  window.crypto.getRandomValues(array);
  return base64URLEncode(array);
}

/**
 * Base64-URL encode (RFC 4648 Section 5)
 * @param {Uint8Array} buffer - Buffer to encode
 * @returns {string} Base64-URL encoded string
 */
function base64URLEncode(buffer) {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}
