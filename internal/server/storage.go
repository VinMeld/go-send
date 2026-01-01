package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VinMeld/go-send/internal/db"
	"github.com/VinMeld/go-send/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// Storage handles persistence for users and files using SQLite3.
type Storage struct {
	DB        *sql.DB
	Queries   *db.Queries
	BlobStore BlobStore
}

// NewStorage creates a new Storage instance.
func NewStorage(baseDir string, blobStore BlobStore) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(baseDir, "gosend.db")
	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Apply Schema
	// For simplicity, we just execute the schema.
	// In production, use a migration tool.
	// We'll use a simple "create if not exists" approach from the schema file.
	// But wait, I need to access the schema file.
	// I'll assume I can embed it or just define it here.
	// Since I can't easily embed a file from a parent directory without go.mod changes or moving files,
	// I'll define the schema const here for now to ensure it works.

	const schema = `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		identity_public_key BLOB NOT NULL,
		exchange_public_key BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		sender TEXT NOT NULL,
		recipient TEXT NOT NULL,
		file_name TEXT NOT NULL,
		encrypted_key BLOB NOT NULL,
		auto_delete BOOLEAN NOT NULL DEFAULT 0,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(sender) REFERENCES users(username),
		FOREIGN KEY(recipient) REFERENCES users(username)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(username) REFERENCES users(username)
	);

	CREATE TABLE IF NOT EXISTS challenges (
		username TEXT PRIMARY KEY,
		nonce TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(username) REFERENCES users(username)
	);
	`

	if _, err := sqliteDB.Exec(schema); err != nil {
		sqliteDB.Close()
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}

	return &Storage{
		DB:        sqliteDB,
		Queries:   db.New(sqliteDB),
		BlobStore: blobStore,
	}, nil
}

// Close closes the database connection.
func (s *Storage) Close() error {
	return s.DB.Close()
}

// AddUser adds or updates a user.
func (s *Storage) AddUser(ctx context.Context, user models.User) error {
	return s.Queries.CreateUser(ctx, db.CreateUserParams{
		Username:          user.Username,
		IdentityPublicKey: user.IdentityPublicKey,
		ExchangePublicKey: user.ExchangePublicKey,
	})
}

// GetUser retrieves a user by username.
func (s *Storage) GetUser(ctx context.Context, username string) (models.User, bool) {
	u, err := s.Queries.GetUser(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, false
		}
		// Log error?
		return models.User{}, false
	}
	return models.User{
		Username:          u.Username,
		IdentityPublicKey: u.IdentityPublicKey,
		ExchangePublicKey: u.ExchangePublicKey,
	}, true
}

// ListAllUsers returns all registered users.
func (s *Storage) ListAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.Queries.ListAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	var result []models.User
	for _, u := range users {
		result = append(result, models.User{
			Username:          u.Username,
			IdentityPublicKey: u.IdentityPublicKey,
			ExchangePublicKey: u.ExchangePublicKey,
		})
	}
	return result, nil
}

// DeleteUser deletes a user from the database.
func (s *Storage) DeleteUser(ctx context.Context, username string) error {
	return s.Queries.DeleteUser(ctx, username)
}

// SaveFile saves a file and its metadata.
func (s *Storage) SaveFile(ctx context.Context, metadata models.FileMetadata, content []byte) error {
	// Save content to BlobStore first
	if err := s.BlobStore.Save(metadata.ID, content); err != nil {
		return err
	}

	// Save metadata to DB
	return s.Queries.CreateFile(ctx, db.CreateFileParams{
		ID:           metadata.ID,
		Sender:       metadata.Sender,
		Recipient:    metadata.Recipient,
		FileName:     metadata.FileName,
		EncryptedKey: metadata.EncryptedKey,
		AutoDelete:   metadata.AutoDelete,
		Timestamp:    metadata.Timestamp,
	})
}

// GetFileMetadata retrieves metadata for a file.
func (s *Storage) GetFileMetadata(ctx context.Context, id string) (models.FileMetadata, bool) {
	f, err := s.Queries.GetFile(ctx, id)
	if err != nil {
		return models.FileMetadata{}, false
	}
	return models.FileMetadata{
		ID:           f.ID,
		Sender:       f.Sender,
		Recipient:    f.Recipient,
		FileName:     f.FileName,
		EncryptedKey: f.EncryptedKey,
		AutoDelete:   f.AutoDelete,
		Timestamp:    f.Timestamp,
	}, true
}

// GetFileContent retrieves the content of a file.
func (s *Storage) GetFileContent(id string) ([]byte, error) {
	return s.BlobStore.Get(id)
}

// ListFiles returns files for a specific recipient.
func (s *Storage) ListFiles(ctx context.Context, recipient string) ([]models.FileMetadata, error) {
	files, err := s.Queries.ListFiles(ctx, recipient)
	if err != nil {
		return nil, err
	}
	var result []models.FileMetadata
	for _, f := range files {
		result = append(result, models.FileMetadata{
			ID:           f.ID,
			Sender:       f.Sender,
			Recipient:    f.Recipient,
			FileName:     f.FileName,
			EncryptedKey: f.EncryptedKey,
			AutoDelete:   f.AutoDelete,
			Timestamp:    f.Timestamp,
		})
	}
	return result, nil
}

// DeleteFile removes a file and its metadata.
func (s *Storage) DeleteFile(ctx context.Context, id string) error {
	// Remove from BlobStore
	if err := s.BlobStore.Delete(id); err != nil {
		return err
	}
	// Remove from DB
	return s.Queries.DeleteFile(ctx, id)
}

// CreateChallenge generates and stores a nonce for a user.
func (s *Storage) CreateChallenge(ctx context.Context, username string, nonce string) error {
	return s.Queries.CreateChallenge(ctx, db.CreateChallengeParams{
		Username: username,
		Nonce:    nonce,
	})
}

// GetChallenge retrieves and deletes a challenge for a user.
func (s *Storage) GetChallenge(ctx context.Context, username string) (string, bool) {
	nonce, err := s.Queries.GetChallenge(ctx, username)
	if err != nil {
		return "", false
	}
	// Delete challenge after retrieval (one-time use)
	_ = s.Queries.DeleteChallenge(ctx, username)
	return nonce, true
}

// CreateSession stores a new session.
func (s *Storage) CreateSession(ctx context.Context, session models.Session) error {
	return s.Queries.CreateSession(ctx, db.CreateSessionParams{
		Token:     session.Token,
		Username:  session.Username,
		ExpiresAt: session.ExpiresAt,
	})
}

// GetSession retrieves a session by token.
func (s *Storage) GetSession(ctx context.Context, token string) (models.Session, bool) {
	sess, err := s.Queries.GetSession(ctx, token)
	if err != nil {
		return models.Session{}, false
	}
	return models.Session{
		Token:     sess.Token,
		Username:  sess.Username,
		ExpiresAt: sess.ExpiresAt,
	}, true
}

// DeleteSession removes a session.
func (s *Storage) DeleteSession(ctx context.Context, token string) error {
	return s.Queries.DeleteSession(ctx, token)
}
