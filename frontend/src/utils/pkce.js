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

/**
 * Base64-URL decode (RFC 4648 Section 5)
 * @param {string} str - Base64-URL encoded string
 * @returns {Uint8Array} Decoded buffer
 */
function base64URLDecode(str) {
  // Convert base64-url to standard base64
  const base64 = str
    .replace(/-/g, '+')
    .replace(/_/g, '/');
  
  // Decode base64
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Open or create IndexedDB database for OAuth keys
 * @returns {Promise<IDBDatabase>} Database instance
 */
async function openKeyDatabase() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('devsmith-oauth', 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains('keys')) {
        db.createObjectStore('keys');
      }
    };
  });
}

/**
 * Get or create encryption key from IndexedDB
 * Key persists across page reloads and tabs
 * @returns {Promise<CryptoKey>} AES-GCM encryption key
 */
async function getOrCreateEncryptionKey() {
  const db = await openKeyDatabase();
  
  // Try to get existing key
  const transaction = db.transaction('keys', 'readonly');
  const store = transaction.objectStore('keys');
  const getRequest = store.get('encryption-key');
  
  const existingKey = await new Promise((resolve) => {
    getRequest.onsuccess = () => resolve(getRequest.result);
    getRequest.onerror = () => resolve(null);
  });
  
  if (existingKey) {
    db.close();
    return existingKey;
  }
  
  // Generate new key
  const newKey = await window.crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    false, // not extractable (cannot be exported)
    ['encrypt', 'decrypt']
  );
  
  // Store in IndexedDB
  const writeTransaction = db.transaction('keys', 'readwrite');
  const writeStore = writeTransaction.objectStore('keys');
  writeStore.put(newKey, 'encryption-key');
  
  await new Promise((resolve, reject) => {
    writeTransaction.oncomplete = () => resolve();
    writeTransaction.onerror = () => reject(writeTransaction.error);
  });
  
  db.close();
  return newKey;
}

/**
 * Encrypt code verifier into self-contained state parameter
 * State contains: verifier + timestamp + random nonce (all encrypted with AES-GCM)
 * @param {string} verifier - PKCE code verifier to encrypt
 * @returns {Promise<string>} Base64-URL encoded encrypted state
 */
export async function encryptVerifier(verifier) {
  const encoder = new TextEncoder();
  
  // Create payload with verifier, timestamp, and random nonce
  const payload = {
    verifier,
    timestamp: Date.now(),
    nonce: base64URLEncode(window.crypto.getRandomValues(new Uint8Array(16)))
  };
  
  const data = encoder.encode(JSON.stringify(payload));
  
  // Get encryption key from IndexedDB
  const key = await getOrCreateEncryptionKey();
  
  // Generate random IV (12 bytes for AES-GCM)
  const iv = window.crypto.getRandomValues(new Uint8Array(12));
  
  // Encrypt
  const encrypted = await window.crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    data
  );
  
  // Combine IV + ciphertext
  const combined = new Uint8Array(iv.length + encrypted.byteLength);
  combined.set(iv, 0);
  combined.set(new Uint8Array(encrypted), iv.length);
  
  // Encode as base64-url
  return base64URLEncode(combined);
}

/**
 * Decrypt verifier from encrypted state parameter
 * Validates timestamp to prevent replay attacks (10 minute expiration)
 * @param {string} encryptedState - Base64-URL encoded encrypted state
 * @returns {Promise<string>} Original code verifier
 * @throws {Error} If state is expired, invalid, or tampered
 */
export async function decryptVerifier(encryptedState) {
  try {
    // Decode base64-url
    const combined = base64URLDecode(encryptedState);
    
    // Extract IV (first 12 bytes) and ciphertext
    const iv = combined.slice(0, 12);
    const ciphertext = combined.slice(12);
    
    // Get encryption key from IndexedDB
    const key = await getOrCreateEncryptionKey();
    
    // Decrypt
    const decrypted = await window.crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      key,
      ciphertext
    );
    
    // Parse JSON payload
    const decoder = new TextDecoder();
    const payload = JSON.parse(decoder.decode(decrypted));
    
    // Validate timestamp (prevent replay > 10 minutes)
    const age = Date.now() - payload.timestamp;
    if (age > 600000) { // 10 minutes in milliseconds
      throw new Error('State expired (>10 minutes old)');
    }
    
    // Return verifier
    return payload.verifier;
  } catch (error) {
    if (error.message === 'State expired (>10 minutes old)') {
      throw error;
    }
    throw new Error('Invalid or tampered state parameter');
  }
}
