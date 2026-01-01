package server

import (
	"os"
	"testing"

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
	if err := s.AddUser(user); err != nil {
		t.Errorf("AddUser failed: %v", err)
	}

	retrieved, ok := s.GetUser("alice")
	if !ok || retrieved.Username != "alice" {
		t.Error("GetUser failed")
	}

	// Test File Operations
	meta := models.FileMetadata{
		ID: "file1", Sender: "alice", Recipient: "bob", FileName: "test.txt",
	}
	content := []byte("hello world")
	if err := s.SaveFile(meta, content); err != nil {
		t.Errorf("SaveFile failed: %v", err)
	}

	files := s.ListFiles("bob")
	if len(files) != 1 || files[0].ID != "file1" {
		t.Error("ListFiles failed")
	}

	retrievedContent, err := s.GetFileContent("file1")
	if err != nil || string(retrievedContent) != string(content) {
		t.Error("GetFileContent failed")
	}

	if err := s.DeleteFile("file1"); err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
	if _, ok := s.GetFileMetadata("file1"); ok {
		t.Error("File should be deleted")
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
