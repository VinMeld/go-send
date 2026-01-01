package models

import "time"

// User represents a user in the system.
type User struct {
	Username          string `json:"username"`
	IdentityPublicKey []byte `json:"identity_public_key"` // Ed25519 public key for signing
	ExchangePublicKey []byte `json:"exchange_public_key"` // X25519 public key for encryption
}

// FileMetadata contains information about an encrypted file.
type FileMetadata struct {
	ID           string    `json:"id"`
	Sender       string    `json:"sender"`
	Recipient    string    `json:"recipient"`
	EncryptedKey []byte    `json:"encrypted_key"` // Symmetric key encrypted with recipient's public key
	Timestamp    time.Time `json:"timestamp"`
	FileName     string    `json:"file_name"` // Original filename
	AutoDelete   bool      `json:"auto_delete"`
}

// UploadRequest is the payload for uploading a file.
type UploadRequest struct {
	Metadata         FileMetadata `json:"metadata"`
	EncryptedContent []byte       `json:"encrypted_content"`
}

// AuthChallenge represents a challenge sent by the server.
type AuthChallenge struct {
	Username string `json:"username"`
	Nonce    string `json:"nonce"`
}

// AuthResponse is the client's response to an authentication challenge.
type AuthResponse struct {
	Username  string `json:"username"`
	Nonce     string `json:"nonce"`
	Signature []byte `json:"signature"`
}

// Session represents an authenticated session.
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}
