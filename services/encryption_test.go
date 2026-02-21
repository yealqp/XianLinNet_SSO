// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"testing"
)

// TestRSAEncryptionDecryption tests RSA encryption and decryption
// Requirements: 9.4
func TestRSAEncryptionDecryption(t *testing.T) {
	// Initialize RSA keys
	err := InitRSAKeys()
	if err != nil {
		t.Fatalf("Failed to initialize RSA keys: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Simple text",
			plaintext: "Hello, World!",
		},
		{
			name:      "Chinese name",
			plaintext: "张三",
		},
		{
			name:      "ID card number",
			plaintext: "110101199001011234",
		},
		{
			name:      "Empty string",
			plaintext: "",
		},
		{
			name:      "Long text",
			plaintext: "This is a longer text that should still be encrypted and decrypted correctly using RSA encryption.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt the plaintext
			ciphertext, err := EncryptData(tt.plaintext)
			if err != nil {
				t.Fatalf("Failed to encrypt data: %v", err)
			}

			// Verify ciphertext is not empty (unless plaintext was empty)
			if tt.plaintext != "" && ciphertext == "" {
				t.Errorf("Expected non-empty ciphertext for non-empty plaintext")
			}

			// Verify ciphertext is different from plaintext
			if tt.plaintext != "" && ciphertext == tt.plaintext {
				t.Errorf("Ciphertext should be different from plaintext")
			}

			// Decrypt the ciphertext
			decrypted, err := DecryptData(ciphertext)
			if err != nil {
				t.Fatalf("Failed to decrypt data: %v", err)
			}

			// Verify decrypted text matches original plaintext
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text does not match original plaintext. Expected: %s, Got: %s", tt.plaintext, decrypted)
			}
		})
	}
}

// TestEncryptDataWithoutInitialization tests encryption without initialization
func TestEncryptDataWithoutInitialization(t *testing.T) {
	// Reset keys to nil to simulate uninitialized state
	originalPublicKey := publicKey
	originalPrivateKey := privateKey
	publicKey = nil
	privateKey = nil

	// Restore keys after test
	defer func() {
		publicKey = originalPublicKey
		privateKey = originalPrivateKey
	}()

	// Try to encrypt without initialization
	_, err := EncryptData("test")
	if err == nil {
		t.Errorf("Expected error when encrypting without initialization, got nil")
	}
}

// TestDecryptDataWithoutInitialization tests decryption without initialization
func TestDecryptDataWithoutInitialization(t *testing.T) {
	// Reset keys to nil to simulate uninitialized state
	originalPublicKey := publicKey
	originalPrivateKey := privateKey
	publicKey = nil
	privateKey = nil

	// Restore keys after test
	defer func() {
		publicKey = originalPublicKey
		privateKey = originalPrivateKey
	}()

	// Try to decrypt without initialization
	_, err := DecryptData("test")
	if err == nil {
		t.Errorf("Expected error when decrypting without initialization, got nil")
	}
}

// TestDecryptInvalidData tests decryption of invalid data
func TestDecryptInvalidData(t *testing.T) {
	// Initialize RSA keys
	err := InitRSAKeys()
	if err != nil {
		t.Fatalf("Failed to initialize RSA keys: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "Invalid base64",
			ciphertext: "not-valid-base64!@#$",
		},
		{
			name:       "Valid base64 but invalid ciphertext",
			ciphertext: "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptData(tt.ciphertext)
			if err == nil {
				t.Errorf("Expected error when decrypting invalid data, got nil")
			}
		})
	}
}

// TestEncryptionConsistency tests that encryption produces different ciphertexts
// for the same plaintext (due to random padding in OAEP)
func TestEncryptionConsistency(t *testing.T) {
	// Initialize RSA keys
	err := InitRSAKeys()
	if err != nil {
		t.Fatalf("Failed to initialize RSA keys: %v", err)
	}

	plaintext := "Test consistency"

	// Encrypt the same plaintext twice
	ciphertext1, err := EncryptData(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt data (first time): %v", err)
	}

	ciphertext2, err := EncryptData(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt data (second time): %v", err)
	}

	// Ciphertexts should be different due to random padding
	if ciphertext1 == ciphertext2 {
		t.Errorf("Expected different ciphertexts for same plaintext due to random padding")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := DecryptData(ciphertext1)
	if err != nil {
		t.Fatalf("Failed to decrypt first ciphertext: %v", err)
	}

	decrypted2, err := DecryptData(ciphertext2)
	if err != nil {
		t.Fatalf("Failed to decrypt second ciphertext: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("Decrypted texts do not match original plaintext")
	}
}
