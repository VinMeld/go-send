package server

import (
	"os"
	"path/filepath"
)

// BlobStore defines the interface for storing file content.
type BlobStore interface {
	Save(id string, content []byte) error
	Get(id string) ([]byte, error)
	Delete(id string) error
}

// LocalBlobStore implements BlobStore using the local filesystem.
type LocalBlobStore struct {
	BaseDir string
}

func NewLocalBlobStore(baseDir string) *LocalBlobStore {
	return &LocalBlobStore{BaseDir: baseDir}
}

func (s *LocalBlobStore) Save(id string, content []byte) error {
	filePath := filepath.Join(s.BaseDir, id+".bin")
	return os.WriteFile(filePath, content, 0644)
}

func (s *LocalBlobStore) Get(id string) ([]byte, error) {
	filePath := filepath.Join(s.BaseDir, id+".bin")
	return os.ReadFile(filePath)
}

func (s *LocalBlobStore) Delete(id string) error {
	filePath := filepath.Join(s.BaseDir, id+".bin")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
