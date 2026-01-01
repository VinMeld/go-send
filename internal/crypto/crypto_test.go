package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateIdentityKeyPair(t *testing.T) {
	kp, err := GenerateIdentityKeyPair()
	if err != nil {
		t.Fatalf("GenerateIdentityKeyPair failed: %v", err)
	}
	if len(kp.Public) != 32 {
		t.Error("Public key should be 32 bytes")
	}
	if len(kp.Private) != 64 { // Ed25519 private key is 64 bytes
		t.Errorf("Private key should be 64 bytes, got %d", len(kp.Private))
	}
}

func TestGenerateExchangeKeyPair(t *testing.T) {
	kp, err := GenerateExchangeKeyPair()
	if err != nil {
		t.Fatalf("GenerateExchangeKeyPair failed: %v", err)
	}
	if kp.Public == nil || len(kp.Public) != 32 {
		t.Error("Public key should be 32 bytes")
	}
	if kp.Private == nil || len(kp.Private) != 32 {
		t.Error("Private key should be 32 bytes")
	}
}

func TestSignVerify(t *testing.T) {
	kp, _ := GenerateIdentityKeyPair()
	message := []byte("Hello, World!")
	sig := Sign(kp.Private, message)
	if !Verify(kp.Public, message, sig) {
		t.Error("Signature verification failed")
	}
	if Verify(kp.Public, []byte("Wrong message"), sig) {
		t.Error("Signature verification should fail for wrong message")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	alice, err := GenerateExchangeKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	bob, err := GenerateExchangeKeyPair()
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
	alice, _ := GenerateExchangeKeyPair()
	bob, _ := GenerateExchangeKeyPair()
	eve, _ := GenerateExchangeKeyPair()

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
