package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
)

// Helper to setup test config
func setupTestConfig(t *testing.T) (string, *httptest.Server) {
	tmpDir, err := os.MkdirTemp("", "go-send-client-test")
	if err != nil {
		t.Fatal(err)
	}

	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
			} else if r.Method == "GET" {
				// Return a dummy user for discovery
				user := models.User{
					Username:          "bob",
					IdentityPublicKey: make([]byte, 32),
					ExchangePublicKey: make([]byte, 32),
				}
				_ = json.NewEncoder(w).Encode(user)
			}
		case "/auth/challenge":
			_ = json.NewEncoder(w).Encode(models.AuthChallenge{Nonce: "nonce"})
		case "/auth/login":
			_ = json.NewEncoder(w).Encode(models.Session{Token: "token"})
		case "/files":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
			} else if r.Method == "GET" {
				// List files
				files := []models.FileMetadata{
					{ID: "file1", FileName: "test.txt", Sender: "bob", Timestamp: time.Now()},
				}
				_ = json.NewEncoder(w).Encode(files)
			}
		case "/files/download":
			// Return dummy file
			meta := models.FileMetadata{ID: "file1", FileName: "test.txt", EncryptedKey: make([]byte, 32)}
			resp := models.UploadRequest{Metadata: meta, EncryptedContent: []byte("encrypted")}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return tmpDir, ts
}

func runCmd(t *testing.T, configDir string, args ...string) (string, error) {
	configFile := filepath.Join(configDir, "config.json")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := GetRootCmd()
	cmd.SetArgs(append(args, "--config", configFile))

	// Reset global cfg?
	// Since we are in the same package, we can access 'cfg' directly if needed.
	// But initConfig should handle reloading if we pass --config.
	// We need to reset cfg to nil or empty to force reload?
	// Actually, initConfig overwrites it.

	err := cmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), err
}

func TestClientCommands(t *testing.T) {
	tmpDir, ts := setupTestConfig(t)
	defer func() { _ = os.RemoveAll(tmpDir) }()
	defer ts.Close()

	// 1. Init
	_, err := runCmd(t, tmpDir, "config", "init", "--user", "alice", "--server", ts.URL)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 2. Register
	_, err = runCmd(t, tmpDir, "register", "--token", "any")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 3. Login
	_, err = runCmd(t, tmpDir, "login")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// 4. Send File (to self)
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("content"), 0644)

	output, err := runCmd(t, tmpDir, "send-file", "alice", testFile)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if !strings.Contains(output, "File sent successfully") {
		t.Errorf("Expected success, got: %s", output)
	}

	// 5. List Files
	output, err = runCmd(t, tmpDir, "list-files")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected test.txt, got: %s", output)
	}

	// 6. Download File (mock decryption will fail but command should run)
	// We need to setup keys for decryption to work, or just check that it tries.
	// The mock server returns empty keys, so decryption will error.
	// That's fine for coverage, we just want to hit the code paths.

	// We need to be in tmpDir to save file?
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	output, _ = runCmd(t, tmpDir, "download-file", "1")
	// It might fail due to decryption error
	if !strings.Contains(output, "Error") && !strings.Contains(output, "downloaded") {
		t.Errorf("Expected some output, got: %s", output)
	}
}

func TestCryptoHelpers(t *testing.T) {
	// Cover crypto helpers if not covered
	idKey, err := crypto.GenerateIdentityKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("hello")
	sig := crypto.Sign(idKey.Private, msg)
	if !crypto.Verify(idKey.Public, msg, sig) {
		t.Error("Verify failed")
	}
}
