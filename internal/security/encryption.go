package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// EncryptionService provides AES-256-GCM encryption for secure storage
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService creates a new encryption service using master key from environment
func NewEncryptionService() (*EncryptionService, error) {
	// Load master key from environment variable
	masterKeyStr := os.Getenv("DEVSMITH_MASTER_KEY")
	if masterKeyStr == "" {
		return nil, fmt.Errorf("DEVSMITH_MASTER_KEY environment variable not set")
	}

	masterKey := []byte(masterKeyStr)

	// Validate key length (AES-256 requires 32 bytes)
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be exactly 32 bytes, got %d bytes", len(masterKey))
	}

	return &EncryptionService{
		masterKey: masterKey,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	// Create cipher block
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate (nonce is prepended to ciphertext)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for safe storage in database
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext
func (s *EncryptionService) Decrypt(ciphertextB64 string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Get nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and actual ciphertext
	nonce, ciphertextOnly := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify authentication
	plaintext, err := gcm.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
