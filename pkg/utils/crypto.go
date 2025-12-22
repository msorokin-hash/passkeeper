package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GeneratePasswordHash generates a bcrypt hash of the given password.
// Uses bcrypt.DefaultCost for the hash complexity.
// Returns the hashed password as a string or an error if hashing fails.
func GeneratePasswordHash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePwdAndHash compares a password with its bcrypt hash to verify if they match.
// Returns true if the password matches the hash, false otherwise.
func ComparePwdAndHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateRandomBytes generates a cryptographically secure random byte array of the specified size.
// Uses crypto/rand for secure random number generation.
// Returns the random bytes or an error if random generation fails.
func GenerateRandomBytes(size int) ([]byte, error) {
	bufBytes := make([]byte, size)
	_, err := rand.Read(bufBytes)
	if err != nil {
		return nil, err
	}

	return bufBytes, nil
}

// GenerateRandomSecretKey generates a random secret key suitable for AES encryption.
// The key size is 2 * aes.BlockSize (32 bytes for AES-256).
// Returns the generated key or an error if random generation fails.
func GenerateRandomSecretKey() ([]byte, error) {
	generatedKey, err := GenerateRandomBytes(2 * aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return generatedKey, nil
}

// Encrypt encrypts data using AES-GCM with the provided hex-encoded key.
// The key must be a valid hex string of appropriate length for AES (typically 32 bytes for AES-256).
// Generates a random nonce and prepends it to the ciphertext.
// Returns the encrypted data (nonce + ciphertext) or an error if encryption fails.
func Encrypt(data []byte, key string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	aesblock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonce, err := GenerateRandomBytes(aesgcm.NonceSize())
	if err != nil {
		return nil, err
	}
	dst := aesgcm.Seal(nonce, nonce, data, nil)

	return dst, nil
}

// Decrypt decrypts data using AES-GCM with the provided hex-encoded key.
// Expects the input data to be in the format: nonce + ciphertext (as produced by Encrypt).
// The key must match the one used for encryption.
// Returns the decrypted plaintext or an error if decryption fails.
func Decrypt(data []byte, key string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	aesblock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
