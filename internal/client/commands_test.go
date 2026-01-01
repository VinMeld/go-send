package client

import (
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
		Users:       make(map[string]models.User),
		PrivateKeys: make(map[string][]byte),
		ServerURL:   "http://localhost:8080",
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

	// configInitCmd is the variable name
	_ = configInitCmd.Flags().Set("user", "alice")

	// Execute Run
	configInitCmd.Run(configInitCmd, []string{})

	if cfg.CurrentUsername != "alice" {
		t.Errorf("Expected current user alice, got %s", cfg.CurrentUsername)
	}
	if _, ok := cfg.PrivateKeys["alice"]; !ok {
		t.Error("Expected private key for alice")
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

	// Add user
	validKey := "dWkdQoVxXQj/ArSR5+It0g/H7dlBq8iB6WQVoDbfm1Q="
	args := []string{"bob", validKey}

	addUserCmd.Run(addUserCmd, args)

	if _, ok := cfg.Users["bob"]; !ok {
		t.Error("Expected bob in users")
	}
}

func TestSetUserCmd(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg.Users["charlie"] = models.User{Username: "charlie"}
	cfg.PrivateKeys["charlie"] = []byte("key")

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
