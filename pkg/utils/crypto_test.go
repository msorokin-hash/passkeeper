package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGeneratePasswordHash(t *testing.T) {
	password := "securepassword123"

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"normal password", password, false},
		{"empty password", "", false},
		{"long password", string(make([]byte, 100)), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := GeneratePasswordHash(tc.password)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("did not expect error but got: %v", err)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tc.password)); err != nil {
				t.Errorf("hash does not match original password: %v", err)
			}
		})
	}
}

func TestComparePwdAndHash(t *testing.T) {
	password := "securepassword123"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password", password, string(hashed), true},
		{"wrong password", "wrongpassword", string(hashed), false},
		{"empty password", "", string(hashed), false},
		{"empty hash", password, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComparePwdAndHash(tc.password, tc.hash)
			if got != tc.want {
				t.Errorf("ComparePwdAndHash(%q, %q) = %v; want %v", tc.password, tc.hash, got, tc.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	t.Run("Encrypt", func(t *testing.T) {
		testData := []byte("Hello, World!")
		testKey := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

		encrypted, err := Encrypt(testData, testKey)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		if len(encrypted) == 0 {
			t.Error("encrypted data should not be empty")
		}

		if len(encrypted) <= len(testData) {
			t.Error("encrypted data should be longer than original data")
		}

		keyBytes, _ := hex.DecodeString(testKey)
		aesblock, _ := aes.NewCipher(keyBytes)
		aesgcm, _ := cipher.NewGCM(aesblock)

		nonceSize := aesgcm.NonceSize()
		if len(encrypted) < nonceSize {
			t.Errorf("encrypted data too short, expected at least %d bytes", nonceSize)
		}

		nonce := encrypted[:nonceSize]
		ciphertext := encrypted[nonceSize:]

		if len(nonce) != nonceSize {
			t.Errorf("nonce size mismatch: expected %d, got %d", nonceSize, len(nonce))
		}

		if len(ciphertext) == 0 {
			t.Error("ciphertext should not be empty")
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("Decrypt", func(t *testing.T) {
		originalData := []byte("Hello, World!")
		testKey := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

		encrypted, err := Encrypt(originalData, testKey)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		decrypted, err := Decrypt(encrypted, testKey)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if string(decrypted) != string(originalData) {
			t.Errorf("decrypted data doesn't match original. expected: %s, Got: %s",
				string(originalData), string(decrypted))
		}
	})
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"zero bytes", 0},
		{"one byte", 1},
		{"small size", 16},
		{"medium size", 256},
		{"large size", 4096},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GenerateRandomBytes(tc.size)
			if err != nil {
				t.Fatalf("generateRandomBytes failed: %v", err)
			}

			if len(result) != tc.size {
				t.Errorf("expected length %d, got %d", tc.size, len(result))
			}
		})
	}
}
