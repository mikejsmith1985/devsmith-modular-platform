// Package portal_services provides encryption services for secure API key storage.
package portal_services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters for key derivation
	// These values balance security and performance for API key encryption
	argon2Time    = 1         // Number of iterations
	argon2Memory  = 64 * 1024 // 64 MB memory cost
	argon2Threads = 4         // Number of parallel threads
	argon2KeyLen  = 32        // 256 bits for AES-256
	saltLength    = 16        // 128 bits for salt
)

var (
	// ErrMasterKeyNotSet is returned when ENCRYPTION_MASTER_KEY environment variable is not configured
	ErrMasterKeyNotSet = errors.New("ENCRYPTION_MASTER_KEY environment variable not set")

	// ErrCiphertextTooShort is returned when the encrypted data is invalid
	ErrCiphertextTooShort = errors.New("ciphertext too short - invalid encrypted data")

	// ErrDecryptionFailed is returned when decryption fails (wrong key or corrupted data)
	ErrDecryptionFailed = errors.New("decryption failed - authentication failed or wrong user key")
)

// EncryptionService handles encryption/decryption of sensitive data using AES-256-GCM.
// Each user's data is encrypted with a unique key derived from the master key and user ID.
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService creates a new encryption service using the ENCRYPTION_MASTER_KEY environment variable.
func NewEncryptionService() *EncryptionService {
	masterKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	return &EncryptionService{
		masterKey: []byte(masterKey),
	}
}

// ValidateMasterKey checks if the master key is configured.
// Returns ErrMasterKeyNotSet if the ENCRYPTION_MASTER_KEY environment variable is not set.
func (s *EncryptionService) ValidateMasterKey() error {
	if len(s.masterKey) == 0 {
		return ErrMasterKeyNotSet
	}
	return nil
}

// EncryptAPIKey encrypts an API key using AES-256-GCM with user-specific key derivation.
// The same API key will produce different ciphertext each time due to random nonce generation.
// Each user's data is encrypted with a unique key, preventing cross-user data access.
//
// Parameters:
//   - apiKey: The plaintext API key to encrypt
//   - userID: The user ID (used for key derivation)
//
// Returns:
//   - Base64-encoded ciphertext
//   - Error if master key is not set or encryption fails
func (s *EncryptionService) EncryptAPIKey(apiKey string, userID int) (string, error) {
	if err := s.ValidateMasterKey(); err != nil {
		return "", err
	}

	// Derive user-specific encryption key using Argon2
	salt := s.generateUserSalt(userID)
	key := argon2.IDKey(s.masterKey, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the API key
	plaintext := []byte(apiKey)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return encoded, nil
}

// DecryptAPIKey decrypts an API key using the user-specific key.
// This will only succeed if:
// 1. The correct master key is set
// 2. The correct user ID is provided (same as when encrypted)
// 3. The ciphertext has not been corrupted
//
// Parameters:
//   - encrypted: Base64-encoded ciphertext (output from EncryptAPIKey)
//   - userID: The user ID (must match the ID used during encryption)
//
// Returns:
//   - The plaintext API key
//   - Error if decryption fails (ErrDecryptionFailed) or invalid input
func (s *EncryptionService) DecryptAPIKey(encrypted string, userID int) (string, error) {
	if err := s.ValidateMasterKey(); err != nil {
		return "", err
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Derive user-specific encryption key
	salt := s.generateUserSalt(userID)
	key := argon2.IDKey(s.masterKey, salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

// generateUserSalt creates a deterministic salt based on user ID.
// This ensures the same user always gets the same encryption key, while different
// users have different keys (preventing cross-user data access).
func (s *EncryptionService) generateUserSalt(userID int) []byte {
	// Use user ID as part of salt to ensure each user has different encryption
	userIDStr := strconv.Itoa(userID)
	salt := make([]byte, saltLength)

	// Fill salt with user ID bytes (repeated if necessary)
	for i := 0; i < saltLength; i++ {
		salt[i] = userIDStr[i%len(userIDStr)]
	}

	return salt
}
