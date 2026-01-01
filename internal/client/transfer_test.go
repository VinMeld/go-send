package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/models"
)

func TestTransferCommands(t *testing.T) {
	// Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
			return
		}
		if r.URL.Path == "/files" {
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				files := []models.FileMetadata{
					{ID: "file1", FileName: "test.txt", Sender: "alice", Recipient: "bob", Timestamp: time.Now()},
				}
				_ = json.NewEncoder(w).Encode(files)
				return
			}
		}
		if r.URL.Path == "/files/download" {
			resp := models.UploadRequest{
				Metadata:         models.FileMetadata{ID: "file1", FileName: "test.txt", EncryptedKey: make([]byte, 32)},
				EncryptedContent: []byte("encrypted_content"),
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup Config
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()
	cfg.ServerURL = server.URL
	cfg.CurrentUsername = "alice"
	cfg.ExchangePrivateKeys["alice"] = make([]byte, 32)
	cfg.IdentityPrivateKeys["alice"] = make([]byte, 64)
	cfg.Users["bob"] = models.User{
		Username:          "bob",
		IdentityPublicKey: make([]byte, 32),
		ExchangePublicKey: make([]byte, 32),
	}

	// Test Ping
	pingCmd.Run(pingCmd, []string{})

	// Test Send File
	tmpFile := filepath.Join(filepath.Dir(configPath), "test.txt")
	_ = os.WriteFile(tmpFile, []byte("content"), 0644)

	sendFileCmd.Run(sendFileCmd, []string{"bob", tmpFile})

	// Test List Files
	listFilesCmd.Run(listFilesCmd, []string{})

	// Test Download File
	downloadFileCmd.Run(downloadFileCmd, []string{"file1"})
}

func TestTransferErrors(t *testing.T) {
	// Setup Server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	configPath, cleanup := setupTestConfig(t)
	defer cleanup()
	cfg.ServerURL = server.URL
	cfg.CurrentUsername = "alice"
	cfg.ExchangePrivateKeys["alice"] = make([]byte, 32)
	cfg.Users["bob"] = models.User{
		Username:          "bob",
		IdentityPublicKey: make([]byte, 32),
		ExchangePublicKey: make([]byte, 32),
	}

	// Test Send File Error
	tmpFile := filepath.Join(filepath.Dir(configPath), "test.txt")
	_ = os.WriteFile(tmpFile, []byte("content"), 0644)

	sendFileCmd.Run(sendFileCmd, []string{"bob", tmpFile})

	// Test List Files Error
	listFilesCmd.Run(listFilesCmd, []string{})

	// Test Download File Error
	downloadFileCmd.Run(downloadFileCmd, []string{"file1"})

	// Test Send File - File Not Found
	sendFileCmd.Run(sendFileCmd, []string{"bob", "non_existent_file"})

	// Test Send File - Unknown Recipient
	sendFileCmd.Run(sendFileCmd, []string{"unknown", tmpFile})

	// Test Send File - No Private Key
	delete(cfg.ExchangePrivateKeys, "alice")
	sendFileCmd.Run(sendFileCmd, []string{"bob", tmpFile})
}

func TestDownloadFile(t *testing.T) {
	// Setup Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/files/download" {
			id := r.URL.Query().Get("id")
			if id == "file1" {
				resp := models.UploadRequest{
					Metadata: models.FileMetadata{
						ID:           "file1",
						FileName:     "test.txt",
						EncryptedKey: make([]byte, 32),
					},
					EncryptedContent: []byte("encrypted"),
				}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	configPath, cleanup := setupTestConfig(t)
	defer cleanup()
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)
	cfg.ServerURL = server.URL
	cfg.CurrentUsername = "alice"
	cfg.ExchangePrivateKeys["alice"] = make([]byte, 32)

	// Test Download - Success
	downloadFileCmd.Run(downloadFileCmd, []string{"file1"})

	// Test Download - Not Found
	downloadFileCmd.Run(downloadFileCmd, []string{"unknown"})
}
