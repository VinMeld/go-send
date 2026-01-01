package models

import "time"

// User represents a user in the system.
type User struct {
	Username  string `json:"username"`
	PublicKey []byte `json:"public_key"` // Ed25519 or X25519 public key
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
