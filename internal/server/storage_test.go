package server

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/models"
)

func TestStorage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-storage-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	blobStore := NewLocalBlobStore(tmpDir)
	s, err := NewStorage(tmpDir, blobStore)
	if err != nil {
		t.Fatal(err)
	}

	// Test User Operations
	user := models.User{
		Username:          "alice",
		IdentityPublicKey: []byte("id-key"),
		ExchangePublicKey: []byte("ex-key"),
	}
	if err := s.AddUser(context.Background(), user); err != nil {
		t.Errorf("AddUser failed: %v", err)
	}

	retrieved, ok := s.GetUser(context.Background(), "alice")
	if !ok || retrieved.Username != "alice" {
		t.Error("GetUser failed")
	}

	// Test File Operations
	meta := models.FileMetadata{
		ID: "file1", Sender: "alice", Recipient: "bob", FileName: "test.txt", EncryptedKey: []byte("key"),
	}
	content := []byte("hello world")
	if err := s.SaveFile(context.Background(), meta, content); err != nil {
		t.Errorf("SaveFile failed: %v", err)
	}

	files, err := s.ListFiles(context.Background(), "bob")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(files) != 1 || files[0].ID != "file1" {
		t.Error("ListFiles returned wrong files")
	}

	retrievedContent, err := s.GetFileContent("file1")
	if err != nil || string(retrievedContent) != string(content) {
		t.Error("GetFileContent failed")
	}

	if err := s.DeleteFile(context.Background(), "file1"); err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
	if _, ok := s.GetFileMetadata(context.Background(), "file1"); ok {
		t.Error("File should be deleted")
	}

	// Test Session Operations
	session := models.Session{
		Token:     "token1",
		Username:  "alice",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.CreateSession(context.Background(), session); err != nil {
		t.Errorf("CreateSession failed: %v", err)
	}

	retrievedSess, ok := s.GetSession(context.Background(), "token1")
	if !ok || retrievedSess.Username != "alice" {
		t.Error("GetSession failed")
	}

	if err := s.DeleteSession(context.Background(), "token1"); err != nil {
		t.Errorf("DeleteSession failed: %v", err)
	}
	if _, ok := s.GetSession(context.Background(), "token1"); ok {
		t.Error("Session should be deleted")
	}
}

func TestStorageErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-storage-error-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test loading from non-existent directory (should work, creates new)
	blobStore := NewLocalBlobStore(tmpDir)
	_, err = NewStorage(tmpDir, blobStore)
	if err != nil {
		t.Errorf("NewStorage failed on new dir: %v", err)
	}
}
