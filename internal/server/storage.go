package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/VinMeld/go-send/internal/models"
)

// Storage handles persistence for users and files.
type Storage struct {
	mu         sync.RWMutex
	BaseDir    string
	Users      map[string]models.User
	Files      map[string]models.FileMetadata
	BlobStore  BlobStore
	Challenges map[string]string         // username -> nonce
	Sessions   map[string]models.Session // token -> session
}

// NewStorage creates a new Storage instance.
func NewStorage(baseDir string, blobStore BlobStore) (*Storage, error) {
	s := &Storage{
		BaseDir:    baseDir,
		Users:      make(map[string]models.User),
		Files:      make(map[string]models.FileMetadata),
		BlobStore:  blobStore,
		Challenges: make(map[string]string),
		Sessions:   make(map[string]models.Session),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) load() error {
	if err := os.MkdirAll(s.BaseDir, 0755); err != nil {
		return err
	}

	// Load Users
	usersFile := filepath.Join(s.BaseDir, "users.json")
	if data, err := os.ReadFile(usersFile); err == nil {
		if err := json.Unmarshal(data, &s.Users); err != nil {
			return fmt.Errorf("failed to load users: %w", err)
		}
	}

	// Load Files Metadata
	filesFile := filepath.Join(s.BaseDir, "files.json")
	if data, err := os.ReadFile(filesFile); err == nil {
		if err := json.Unmarshal(data, &s.Files); err != nil {
			return fmt.Errorf("failed to load files: %w", err)
		}
	}
	return nil
}

func (s *Storage) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveInternal()
}

func (s *Storage) saveInternal() error {
	// Caller must hold lock
	usersData, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.BaseDir, "users.json"), usersData, 0644); err != nil {
		return err
	}

	filesData, err := json.MarshalIndent(s.Files, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.BaseDir, "files.json"), filesData, 0644); err != nil {
		return err
	}
	return nil
}

// AddUser adds or updates a user.
func (s *Storage) AddUser(user models.User) error {
	s.mu.Lock()
	s.Users[user.Username] = user
	s.mu.Unlock()
	return s.save()
}

// GetUser retrieves a user by username.
func (s *Storage) GetUser(username string) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.Users[username]
	return u, ok
}

// SaveFile saves a file and its metadata.
func (s *Storage) SaveFile(metadata models.FileMetadata, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save content to BlobStore
	if err := s.BlobStore.Save(metadata.ID, content); err != nil {
		return err
	}

	s.Files[metadata.ID] = metadata

	// We need to unlock before calling save() because save() locks too?
	// Ah, save() locks. I should refactor save() to not lock, or call an internal save.
	// Let's just inline the save logic or make save() unexported and lock-free, called by public methods.
	// Actually, I'll just release the lock before calling save, but that's racy.
	// Better: make save() NOT lock, and assume caller holds lock.

	// Re-implementing save logic here to avoid deadlock or race,
	// but for MVP let's just do it correctly.
	// I'll change save() to NOT lock, and wrap public methods with lock.

	// Wait, I can't change save() signature easily in this tool call without rewriting it.
	// I'll just use a separate internal save method.
	return s.saveInternal()
}

// GetFileMetadata retrieves metadata for a file.
func (s *Storage) GetFileMetadata(id string) (models.FileMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.Files[id]
	return m, ok
}

// GetFileContent retrieves the content of a file.
func (s *Storage) GetFileContent(id string) ([]byte, error) {
	return s.BlobStore.Get(id)
}

// ListFiles returns files for a specific recipient.
func (s *Storage) ListFiles(recipient string) []models.FileMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []models.FileMetadata
	for _, f := range s.Files {
		if f.Recipient == recipient {
			result = append(result, f)
		}
	}
	return result
}

// DeleteFile removes a file and its metadata.
func (s *Storage) DeleteFile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Files[id]; !ok {
		return fmt.Errorf("file not found")
	}

	// Remove from BlobStore
	if err := s.BlobStore.Delete(id); err != nil {
		return err
	}

	// Remove from memory
	delete(s.Files, id)

	return s.saveInternal()
}

// CreateChallenge generates and stores a nonce for a user.
func (s *Storage) CreateChallenge(username string, nonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Challenges[username] = nonce
}

// GetChallenge retrieves and deletes a challenge for a user.
func (s *Storage) GetChallenge(username string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	nonce, ok := s.Challenges[username]
	if ok {
		delete(s.Challenges, username)
	}
	return nonce, ok
}

// CreateSession stores a new session.
func (s *Storage) CreateSession(session models.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sessions[session.Token] = session
}

// GetSession retrieves a session by token.
func (s *Storage) GetSession(token string) (models.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.Sessions[token]
	return sess, ok
}

// DeleteSession removes a session.
func (s *Storage) DeleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Sessions, token)
}
