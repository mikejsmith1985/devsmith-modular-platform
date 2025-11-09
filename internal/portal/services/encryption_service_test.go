package portal_services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionService_EncryptDecrypt_RoundTrip(t *testing.T) {
	// Setup: Set master key
	os.Setenv("ENCRYPTION_MASTER_KEY", "test-master-key-32-bytes-long!")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	// Test data
	apiKey := "sk-test-api-key-1234567890"
	userID := 123

	// Encrypt
	encrypted, err := service.EncryptAPIKey(apiKey, userID)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, apiKey, encrypted, "Encrypted should differ from plaintext")

	// Decrypt
	decrypted, err := service.DecryptAPIKey(encrypted, userID)
	require.NoError(t, err)
	assert.Equal(t, apiKey, decrypted, "Decrypted should match original")
}

func TestEncryptionService_EncryptAPIKey_NonceRandomness(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "test-master-key-32-bytes-long!")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	apiKey := "sk-same-key"
	userID := 456

	// Encrypt same key twice
	encrypted1, err := service.EncryptAPIKey(apiKey, userID)
	require.NoError(t, err)

	encrypted2, err := service.EncryptAPIKey(apiKey, userID)
	require.NoError(t, err)

	// Ciphertexts should differ due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2, "Same plaintext should produce different ciphertext due to nonce")
}

func TestEncryptionService_DecryptAPIKey_WrongUserID(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "test-master-key-32-bytes-long!")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	apiKey := "sk-test-key"
	userID := 789

	// Encrypt with userID 789
	encrypted, err := service.EncryptAPIKey(apiKey, userID)
	require.NoError(t, err)

	// Try to decrypt with different userID
	_, err = service.DecryptAPIKey(encrypted, 999)
	assert.Error(t, err, "Decryption should fail with wrong user ID")
	assert.Contains(t, err.Error(), "authentication failed", "Error should indicate auth failure")
}

func TestEncryptionService_DecryptAPIKey_CorruptedCiphertext(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "test-master-key-32-bytes-long!")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	corruptedData := "this-is-not-valid-base64-encrypted-data!!!"
	userID := 111

	_, err := service.DecryptAPIKey(corruptedData, userID)
	assert.Error(t, err, "Should fail to decrypt corrupted data")
}

func TestEncryptionService_ValidateMasterKey_Present(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "valid-key-present")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	err := service.ValidateMasterKey()
	assert.NoError(t, err, "Should validate when master key is present")
}

func TestEncryptionService_ValidateMasterKey_Missing(t *testing.T) {
	os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	err := service.ValidateMasterKey()
	assert.Error(t, err, "Should fail when master key is missing")
	assert.Contains(t, err.Error(), "ENCRYPTION_MASTER_KEY", "Error should mention the missing env var")
}

func TestEncryptionService_UserSpecificSalt(t *testing.T) {
	os.Setenv("ENCRYPTION_MASTER_KEY", "test-master-key-32-bytes-long!")
	defer os.Unsetenv("ENCRYPTION_MASTER_KEY")

	service := NewEncryptionService()

	apiKey := "sk-same-key-for-both-users"

	// Encrypt for user 1
	encrypted1, err := service.EncryptAPIKey(apiKey, 1)
	require.NoError(t, err)

	// Encrypt for user 2 (same key)
	encrypted2, err := service.EncryptAPIKey(apiKey, 2)
	require.NoError(t, err)

	// Ciphertexts should differ (different derived keys per user)
	assert.NotEqual(t, encrypted1, encrypted2, "Different users should produce different ciphertext")

	// User 1 can decrypt their own
	decrypted1, err := service.DecryptAPIKey(encrypted1, 1)
	require.NoError(t, err)
	assert.Equal(t, apiKey, decrypted1)

	// User 2 can decrypt their own
	decrypted2, err := service.DecryptAPIKey(encrypted2, 2)
	require.NoError(t, err)
	assert.Equal(t, apiKey, decrypted2)

	// User 1 CANNOT decrypt user 2's data
	_, err = service.DecryptAPIKey(encrypted2, 1)
	assert.Error(t, err, "User 1 should not decrypt user 2's data")
}
