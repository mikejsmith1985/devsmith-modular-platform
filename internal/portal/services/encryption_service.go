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
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32 // 256 bits for AES-256
	saltLength    = 16 // 128 bits
)

// EncryptionService handles encryption/decryption of sensitive data
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService() *EncryptionService {
	masterKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	return &EncryptionService{
		masterKey: []byte(masterKey),
	}
}

// ValidateMasterKey checks if the master key is configured
func (s *EncryptionService) ValidateMasterKey() error {
	if len(s.masterKey) == 0 {
		return errors.New("ENCRYPTION_MASTER_KEY environment variable not set")
	}
	return nil
}

// EncryptAPIKey encrypts an API key using AES-256-GCM with user-specific key derivation
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

// DecryptAPIKey decrypts an API key using the user-specific key
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
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	return string(plaintext), nil
}

// generateUserSalt creates a deterministic salt based on user ID
// This ensures the same user always gets the same encryption key
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
