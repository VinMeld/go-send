package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}
	if len(kp.Public) != 32 || len(kp.Private) != 32 {
		t.Error("Key lengths should be 32 bytes")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	alice, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	bob, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello, Bob! This is a secret.")

	// Alice encrypts for Bob
	encrypted, err := Encrypt(message, bob.Public, alice.Private)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Bob decrypts from Alice
	decrypted, err := Decrypt(encrypted, alice.Public, bob.Private)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(message, decrypted) {
		t.Errorf("Decrypted message does not match original.\nGot: %s\nWant: %s", decrypted, message)
	}
}

func TestDecryptFailure(t *testing.T) {
	alice, _ := GenerateKeyPair()
	bob, _ := GenerateKeyPair()
	eve, _ := GenerateKeyPair()

	message := []byte("Secret")
	encrypted, _ := Encrypt(message, bob.Public, alice.Private)

	// Eve tries to decrypt
	_, err := Decrypt(encrypted, alice.Public, eve.Private)
	if err == nil {
		t.Error("Expected decryption failure for wrong private key, got nil")
	}
}

func TestGenerateSymmetricKey(t *testing.T) {
	key, err := GenerateSymmetricKey()
	if err != nil {
		t.Fatalf("GenerateSymmetricKey failed: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}
}
