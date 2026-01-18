package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var (
	// ErrEncryptionKeyNotSet is returned when the encryption key environment variable is not set
	ErrEncryptionKeyNotSet = errors.New("encryption key not set: ENCRYPTION_KEY environment variable is required")

	// ErrInvalidKeyLength is returned when the encryption key is not the correct length
	ErrInvalidKeyLength = errors.New("invalid encryption key length: must be 32 bytes for AES-256")

	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed: invalid ciphertext or key")

	// ErrCiphertextTooShort is returned when the ciphertext is too short
	ErrCiphertextTooShort = errors.New("ciphertext too short")
)

// EncryptionService provides AES-256-GCM encryption and decryption
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service using the ENCRYPTION_KEY environment variable
func NewEncryptionService() (*EncryptionService, error) {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		return nil, ErrEncryptionKeyNotSet
	}

	// Decode the key from base64 (allows for 32 random bytes encoded as base64)
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		// If not base64, try using the raw string (must be exactly 32 bytes)
		key = []byte(keyStr)
	}

	// Ensure key is exactly 32 bytes for AES-256
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	return &EncryptionService{key: key}, nil
}

// NewEncryptionServiceWithKey creates a new encryption service with a provided key
// The key must be exactly 32 bytes for AES-256
func NewEncryptionServiceWithKey(key []byte) (*EncryptionService, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	return &EncryptionService{key: key}, nil
}

// EncryptValue encrypts a plaintext string using AES-256-GCM and returns a base64-encoded ciphertext
func (s *EncryptionService) EncryptValue(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a unique nonce for each encryption
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext and prepend the nonce
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return as base64-encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue decrypts a base64-encoded ciphertext using AES-256-GCM and returns the plaintext
func (s *EncryptionService) DecryptValue(encodedCiphertext string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	// Extract the nonce and actual ciphertext
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts a byte slice using AES-256-GCM
func (s *EncryptionService) EncryptBytes(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptBytes decrypts a byte slice using AES-256-GCM
func (s *EncryptionService) DecryptBytes(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// GenerateRandomKey generates a new random 32-byte key suitable for AES-256
// Returns the key as a base64-encoded string
func GenerateRandomKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
