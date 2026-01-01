package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/VinMeld/go-send/internal/models"
)

func setupTestConfig(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "go-send-client-cmd-test")
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(tmpDir, "config.json")
	cfg = &Config{
		Users:               make(map[string]models.User),
		IdentityPrivateKeys: make(map[string][]byte),
		ExchangePrivateKeys: make(map[string][]byte),
		SessionTokens:       make(map[string]string),
		ServerURL:           "http://localhost:8080",
	}
	cfgFile = configPath // Set global cfgFile

	return configPath, func() {
		_ = os.RemoveAll(tmpDir)
		cfg = nil
		cfgFile = ""
	}
}

func TestInitCmd(t *testing.T) {
	configPath, cleanup := setupTestConfig(t)
	defer cleanup()

	_ = configInitCmd.Flags().Set("user", "alice")

	// Execute Run
	configInitCmd.Run(configInitCmd, []string{})

	if cfg.CurrentUsername != "alice" {
		t.Errorf("Expected current user alice, got %s", cfg.CurrentUsername)
	}
	if _, ok := cfg.IdentityPrivateKeys["alice"]; !ok {
		t.Error("Expected identity private key for alice")
	}
	if _, ok := cfg.ExchangePrivateKeys["alice"]; !ok {
		t.Error("Expected exchange private key for alice")
	}

	// Verify file was saved
	loaded, _ := LoadConfig(configPath)
	if loaded.CurrentUsername != "alice" {
		t.Error("Config not saved to disk")
	}
}

func TestAddUserCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Add user (expects 3 args: username, id_pub, ex_pub)
	validIdKey := "dWkdQoVxXQj/ArSR5+It0g/H7dlBq8iB6WQVoDbfm1Q="
	validExKey := "dWkdQoVxXQj/ArSR5+It0g/H7dlBq8iB6WQVoDbfm1Q="
	args := []string{"bob", validIdKey, validExKey}

	addUserCmd.Run(addUserCmd, args)

	if _, ok := cfg.Users["bob"]; !ok {
		t.Error("Expected bob in users")
	}
}

func TestSetUserCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg.Users["charlie"] = models.User{Username: "charlie"}
	cfg.IdentityPrivateKeys["charlie"] = make([]byte, 64)

	setUserCmd.Run(setUserCmd, []string{"charlie"})

	if cfg.CurrentUsername != "charlie" {
		t.Errorf("Expected current user charlie, got %s", cfg.CurrentUsername)
	}
}

func TestListUsersCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Just run it to ensure no panic
	listUsersCmd.Run(listUsersCmd, []string{})
}

func TestSetServerCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	setServerCmd.Run(setServerCmd, []string{"http://example.com"})

	if cfg.ServerURL != "http://example.com" {
		t.Errorf("Expected server URL http://example.com, got %s", cfg.ServerURL)
	}
}

func TestRemoveUserCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg.Users["dave"] = models.User{Username: "dave"}
	removeUserCmd.Run(removeUserCmd, []string{"dave"})

	if _, ok := cfg.Users["dave"]; ok {
		t.Error("Expected dave to be removed")
	}
}

func TestRegisterCmd(t *testing.T) {
	// Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" && r.Method == http.MethodPost {
			token := r.Header.Get("X-Registration-Token")
			if token != "secret" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, cleanup := setupTestConfig(t)
	defer cleanup()
	cfg.ServerURL = server.URL

	// Setup User
	cfg.CurrentUsername = "alice"
	cfg.IdentityPrivateKeys["alice"] = make([]byte, 64)
	cfg.ExchangePrivateKeys["alice"] = make([]byte, 32)
	cfg.Users["alice"] = models.User{Username: "alice"}

	// Test Register Success
	_ = registerCmd.Flags().Set("token", "secret")
	registerCmd.Run(registerCmd, []string{})

	// Test Register Failure (Wrong Token)
	_ = registerCmd.Flags().Set("token", "wrong")
	registerCmd.Run(registerCmd, []string{})
}

func TestDeleteFileCmd(t *testing.T) {
	// Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/files" && r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			if id == "file1" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	_, cleanup := setupTestConfig(t)
	defer cleanup()
	cfg.ServerURL = server.URL
	cfg.CurrentUsername = "alice"
	cfg.SessionTokens["alice"] = "valid-token"

	// Test Delete Success
	deleteFileCmd.Run(deleteFileCmd, []string{"file1"})

	// Test Delete Not Found
	deleteFileCmd.Run(deleteFileCmd, []string{"unknown"})
}
