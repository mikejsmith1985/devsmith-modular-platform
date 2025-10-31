package security

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEncryptionService_NewEncryptionService_RequiresMasterKey verifies master key is required
func TestEncryptionService_NewEncryptionService_RequiresMasterKey(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("DEVSMITH_MASTER_KEY")
	defer os.Setenv("DEVSMITH_MASTER_KEY", originalKey)

	// Clear the env var
	os.Unsetenv("DEVSMITH_MASTER_KEY")

	svc, err := NewEncryptionService()
	assert.Error(t, err, "Should error when DEVSMITH_MASTER_KEY not set")
	assert.Nil(t, svc, "Service should be nil on error")
	assert.Contains(t, err.Error(), "DEVSMITH_MASTER_KEY", "Error should mention the env var")
}

// TestEncryptionService_NewEncryptionService_ValidatesMasterKeyLength verifies key length is 32 bytes
func TestEncryptionService_NewEncryptionService_ValidatesMasterKeyLength(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"Too short (16 bytes)", 16, true},
		{"Too short (24 bytes)", 24, true},
		{"Correct (32 bytes)", 32, false},
		{"Too long (48 bytes)", 48, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env var for this subtest
			originalKey := os.Getenv("DEVSMITH_MASTER_KEY")
			defer os.Setenv("DEVSMITH_MASTER_KEY", originalKey)

			// Create a key of specific length
			key := make([]byte, tt.keyLen)
			for i := range key {
				key[i] = 'a' // Fill with printable characters
			}
			os.Setenv("DEVSMITH_MASTER_KEY", string(key))

			svc, err := NewEncryptionService()
			if tt.wantErr {
				assert.Error(t, err, "Should error for incorrect key length")
				assert.Nil(t, svc, "Service should be nil on error")
				assert.Contains(t, err.Error(), "32 bytes", "Error should mention correct key size")
			} else {
				assert.NoError(t, err, "Should not error for correct key length")
				assert.NotNil(t, svc, "Service should be created")
			}
		})
	}
}

// TestEncryptionService_Encrypt_ProducesNonEmptyCiphertext verifies encryption works
func TestEncryptionService_Encrypt_ProducesNonEmptyCiphertext(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)
	require.NotNil(t, svc)

	plaintext := "my-secret-api-key"
	ciphertext, err := svc.Encrypt(plaintext)

	assert.NoError(t, err, "Encryption should succeed")
	assert.NotEmpty(t, ciphertext, "Ciphertext should not be empty")
	assert.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")
}

// TestEncryptionService_Encrypt_IsNonDeterministic verifies encryption uses random nonce
func TestEncryptionService_Encrypt_IsNonDeterministic(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	plaintext := "same-input"

	ciphertext1, err := svc.Encrypt(plaintext)
	require.NoError(t, err)

	ciphertext2, err := svc.Encrypt(plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, ciphertext1, ciphertext2, "Same plaintext should produce different ciphertexts (random nonce)")
}

// TestEncryptionService_Encrypt_HandlesEmptyString verifies empty strings can be encrypted
func TestEncryptionService_Encrypt_HandlesEmptyString(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	ciphertext, err := svc.Encrypt("")

	assert.NoError(t, err, "Should encrypt empty string")
	assert.NotEqual(t, "", ciphertext, "Should produce non-empty ciphertext for empty plaintext")
}

// TestEncryptionService_Decrypt_RecoversOriginalText verifies decryption works
func TestEncryptionService_Decrypt_RecoversOriginalText(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	testCases := []string{
		"simple-key",
		"very-long-api-key-with-many-characters-and-special-symbols-!@#$%",
		"key-with-unicode-™-symbols-中文",
		"",
	}

	for _, plaintext := range testCases {
		t.Run(plaintext, func(t *testing.T) {
			ciphertext, err := svc.Encrypt(plaintext)
			require.NoError(t, err)

			decrypted, err := svc.Decrypt(ciphertext)
			assert.NoError(t, err, "Decryption should succeed")
			assert.Equal(t, plaintext, decrypted, "Decrypted text should match original")
		})
	}
}

// TestEncryptionService_Decrypt_InvalidBase64ReturnsError verifies error handling
func TestEncryptionService_Decrypt_InvalidBase64ReturnsError(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	invalidCiphertext := "not-valid-base64!@#$%"
	decrypted, err := svc.Decrypt(invalidCiphertext)

	assert.Error(t, err, "Should error on invalid base64")
	assert.Empty(t, decrypted, "Should return empty string on error")
}

// TestEncryptionService_Decrypt_TamperedCiphertextReturnsError verifies authentication
func TestEncryptionService_Decrypt_TamperedCiphertextReturnsError(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	plaintext := "sensitive-data"
	ciphertext, err := svc.Encrypt(plaintext)
	require.NoError(t, err)

	// Tamper with the ciphertext (change first character)
	if ciphertext != "" {
		ciphertextBytes := []byte(ciphertext)
		if ciphertextBytes[0] == 'a' {
			ciphertextBytes[0] = 'b'
		} else {
			ciphertextBytes[0] = 'a'
		}
		tamperedCiphertext := string(ciphertextBytes)

		decrypted, err := svc.Decrypt(tamperedCiphertext)
		assert.Error(t, err, "Should error on tampered ciphertext (GCM authentication fails)")
		assert.Empty(t, decrypted, "Should return empty string on auth failure")
		assert.Contains(t, err.Error(), "decryption failed", "Error should indicate auth failure")
	}
}

// TestEncryptionService_Decrypt_ShortCiphertextReturnsError verifies length validation
func TestEncryptionService_Decrypt_ShortCiphertextReturnsError(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	// A valid base64 string but too short to be valid GCM ciphertext
	shortCiphertext := "AAAA" // Decodes to 3 bytes, GCM nonce is at least 12 bytes

	decrypted, err := svc.Decrypt(shortCiphertext)
	assert.Error(t, err, "Should error on too-short ciphertext")
	assert.Equal(t, "", decrypted, "Should return empty string on error")
	assert.Contains(t, err.Error(), "ciphertext too short", "Error should indicate length issue")
}

// TestEncryptionService_RoundTrip_LargeData verifies large data works
func TestEncryptionService_RoundTrip_LargeData(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	// Create a large plaintext (1MB)
	largePlaintext := ""
	for i := 0; i < 1024*100; i++ { // 100K iterations
		largePlaintext += "This is a line of text that will be repeated many times to test large data. "
	}

	ciphertext, err := svc.Encrypt(largePlaintext)
	require.NoError(t, err)

	decrypted, err := svc.Decrypt(ciphertext)
	assert.NoError(t, err, "Decryption of large data should succeed")
	assert.Equal(t, largePlaintext, decrypted, "Decrypted large data should match original")
}

// TestEncryptionService_EncryptDecrypt_DifferentKeysCannotDecrypt verifies key isolation
func TestEncryptionService_EncryptDecrypt_DifferentKeysCannotDecrypt(t *testing.T) {
	svc1, err := setupTestEncryptionServiceWithKey([]byte("12345678901234567890123456789012"))
	require.NoError(t, err)

	svc2, err := setupTestEncryptionServiceWithKey([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	require.NoError(t, err)

	plaintext := "secret-data"
	ciphertext, err := svc1.Encrypt(plaintext)
	require.NoError(t, err)

	// Try to decrypt with different service (different key)
	decrypted, err := svc2.Decrypt(ciphertext)
	assert.Error(t, err, "Should not decrypt with different key")
	assert.Empty(t, decrypted, "Should return empty on auth failure")
}

// TestEncryptionService_Encrypt_HandlesSpecialCharacters verifies non-ASCII data
func TestEncryptionService_Encrypt_HandlesSpecialCharacters(t *testing.T) {
	svc, err := setupTestEncryptionService()
	require.NoError(t, err)

	specialChars := `!@#$%^&*()_+-={}[]|:;"'<>,.?/~`
	ciphertext, err := svc.Encrypt(specialChars)
	require.NoError(t, err)

	decrypted, err := svc.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, specialChars, decrypted, "Special characters should survive encryption/decryption")
}

// Helper function to set up test encryption service with default key
func setupTestEncryptionService() (*EncryptionService, error) {
	// Set a 32-byte test key
	testKey := "12345678901234567890123456789012" // Exactly 32 chars = 32 bytes
	os.Setenv("DEVSMITH_MASTER_KEY", testKey)
	defer os.Unsetenv("DEVSMITH_MASTER_KEY")

	return NewEncryptionService()
}

// Helper function to set up test encryption service with specific key
func setupTestEncryptionServiceWithKey(key []byte) (*EncryptionService, error) {
	os.Setenv("DEVSMITH_MASTER_KEY", string(key))
	defer os.Unsetenv("DEVSMITH_MASTER_KEY")

	return NewEncryptionService()
}
