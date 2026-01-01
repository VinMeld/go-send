package crypto

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

// KeyPair represents a public/private key pair.
type KeyPair struct {
	Public  *[32]byte
	Private *[32]byte
}

// GenerateKeyPair generates a new random key pair for Box.
func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyPair{Public: pub, Private: priv}, nil
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
