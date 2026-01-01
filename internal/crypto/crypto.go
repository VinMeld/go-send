package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// IdentityKeyPair represents an Ed25519 key pair for signing.
type IdentityKeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

// GenerateIdentityKeyPair generates a new Ed25519 key pair.
func GenerateIdentityKeyPair() (*IdentityKeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &IdentityKeyPair{Public: pub, Private: priv}, nil
}

// ExchangeKeyPair represents an X25519 key pair for encryption.
type ExchangeKeyPair struct {
	Public  *[32]byte
	Private *[32]byte
}

// GenerateExchangeKeyPair generates a new X25519 key pair for Box.
func GenerateExchangeKeyPair() (*ExchangeKeyPair, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ExchangeKeyPair{Public: pub, Private: priv}, nil
}

// Sign signs a message with an Ed25519 private key.
func Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// Verify verifies a signature with an Ed25519 public key.
func Verify(publicKey ed25519.PublicKey, message []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// Encrypt encrypts a message for a recipient using their public key and the sender's private key.
// It returns the nonce appended to the ciphertext.
func Encrypt(message []byte, recipientPub *[32]byte, senderPriv *[32]byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	encrypted := box.Seal(nonce[:], message, &nonce, recipientPub, senderPriv)
	return encrypted, nil
}

// Decrypt decrypts a message from a sender using their public key and the recipient's private key.
// It expects the nonce to be prepended to the ciphertext.
func Decrypt(encrypted []byte, senderPub *[32]byte, recipientPriv *[32]byte) ([]byte, error) {
	if len(encrypted) < 24 {
		return nil, errors.New("message too short")
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	ciphertext := encrypted[24:]

	decrypted, ok := box.Open(nil, ciphertext, &nonce, senderPub, recipientPriv)
	if !ok {
		return nil, errors.New("decryption failed")
	}
	return decrypted, nil
}

// GenerateSymmetricKey generates a random 32-byte key.
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
