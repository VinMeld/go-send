package server

import (
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-server-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test Local Storage
	_ = os.Setenv("STORAGE_TYPE", "local")
	srv, err := NewServer(":8081", tmpDir)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if srv.Port != ":8081" {
		t.Errorf("Expected port :8081, got %s", srv.Port)
	}

	// Test S3 Storage (Missing Bucket)
	_ = os.Setenv("STORAGE_TYPE", "s3")
	_ = os.Unsetenv("AWS_BUCKET")
	_, err = NewServer(":8081", tmpDir)
	if err == nil {
		t.Error("Expected error for missing AWS_BUCKET")
	}

	// Reset env
	_ = os.Unsetenv("STORAGE_TYPE")
}
