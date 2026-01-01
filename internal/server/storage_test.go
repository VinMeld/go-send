package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/models"
)

func TestStorage(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "go-send-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	blobStore := NewLocalBlobStore(tmpDir)
	store, err := NewStorage(tmpDir, blobStore)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}

	// Test AddUser
	user := models.User{Username: "alice", PublicKey: []byte("alice_pub_key")}
	if err := store.AddUser(user); err != nil {
		t.Errorf("AddUser failed: %v", err)
	}

	retrievedUser, ok := store.GetUser("alice")
	if !ok {
		t.Error("GetUser failed: user not found")
	}
	if retrievedUser.Username != "alice" {
		t.Errorf("Expected username alice, got %s", retrievedUser.Username)
	}

	// Test SaveFile
	fileID := "file1"
	meta := models.FileMetadata{
		ID:        fileID,
		Sender:    "alice",
		Recipient: "bob",
		Timestamp: time.Now(),
		FileName:  "test.txt",
	}
	content := []byte("file content")

	if err := store.SaveFile(meta, content); err != nil {
		t.Errorf("SaveFile failed: %v", err)
	}

	// Test GetFileMetadata
	retrievedMeta, ok := store.GetFileMetadata(fileID)
	if !ok {
		t.Error("GetFileMetadata failed: file not found")
	}
	if retrievedMeta.FileName != "test.txt" {
		t.Errorf("Expected filename test.txt, got %s", retrievedMeta.FileName)
	}

	// Test GetFileContent
	retrievedContent, err := store.GetFileContent(fileID)
	if err != nil {
		t.Errorf("GetFileContent failed: %v", err)
	}
	if string(retrievedContent) != string(content) {
		t.Errorf("Content mismatch")
	}

	// Test ListFiles
	files := store.ListFiles("bob")
	if len(files) != 1 {
		t.Errorf("Expected 1 file for bob, got %d", len(files))
	}

	// Test DeleteFile
	if err := store.DeleteFile(fileID); err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}

	if _, ok := store.GetFileMetadata(fileID); ok {
		t.Error("File metadata should be gone after delete")
	}

	// Check if file on disk is gone
	if _, err := os.Stat(filepath.Join(tmpDir, fileID+".bin")); !os.IsNotExist(err) {
		t.Error("File content should be deleted from disk")
	}
}

func TestStorageErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-storage-error-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	blobStore := NewLocalBlobStore(tmpDir)
	store, err := NewStorage(tmpDir, blobStore)
	if err != nil {
		t.Fatal(err)
	}

	// Test Load Corrupted Users
	os.WriteFile(filepath.Join(tmpDir, "users.json"), []byte("bad json"), 0644)
	if err := store.load(); err == nil {
		t.Error("Expected error loading corrupted users")
	}

	// Test Load Corrupted Files
	os.WriteFile(filepath.Join(tmpDir, "users.json"), []byte("{}"), 0644) // Fix users
	os.WriteFile(filepath.Join(tmpDir, "files.json"), []byte("bad json"), 0644)
	if err := store.load(); err == nil {
		t.Error("Expected error loading corrupted files")
	}
}
