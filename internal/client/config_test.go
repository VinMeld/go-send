package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/VinMeld/go-send/internal/models"
)

func TestConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-client-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "config.json")

	// Test LoadConfig (New)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.ServerURL == "" {
		t.Error("Expected default server URL")
	}

	// Test SaveConfig
	cfg.CurrentUsername = "alice"
	cfg.Users["bob"] = models.User{
		Username:          "bob",
		IdentityPublicKey: []byte("bob_id_key"),
		ExchangePublicKey: []byte("bob_ex_key"),
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Test LoadConfig (Existing)
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig existing failed: %v", err)
	}
	if loadedCfg.CurrentUsername != "alice" {
		t.Errorf("Expected username alice, got %s", loadedCfg.CurrentUsername)
	}
	if _, ok := loadedCfg.Users["bob"]; !ok {
		t.Error("Expected bob in users")
	}
}
