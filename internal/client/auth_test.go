package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
)

func TestLogin(t *testing.T) {
	// Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/challenge" {
			_ = json.NewEncoder(w).Encode(models.AuthChallenge{
				Username: "alice",
				Nonce:    "test-nonce",
			})
			return
		}
		if r.URL.Path == "/auth/login" {
			var resp models.AuthResponse
			_ = json.NewDecoder(r.Body).Decode(&resp)
			if resp.Nonce != "test-nonce" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(models.Session{
				Token:     "session-token",
				Username:  "alice",
				ExpiresAt: time.Now().Add(1 * time.Hour),
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup Config
	tmpDir, err := os.MkdirTemp("", "go-send-client-auth-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "config.json")
	cfg = &Config{
		Users:               make(map[string]models.User),
		IdentityPrivateKeys: make(map[string][]byte),
		SessionTokens:       make(map[string]string),
		ServerURL:           server.URL,
	}
	cfgFile = configPath

	// Setup User Keys
	idKey, _ := crypto.GenerateIdentityKeyPair()
	cfg.CurrentUsername = "alice"
	cfg.IdentityPrivateKeys["alice"] = idKey.Private
	cfg.Users["alice"] = models.User{Username: "alice", IdentityPublicKey: idKey.Public}

	// Test Login
	if err := Login(); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if cfg.SessionTokens["alice"] != "session-token" {
		t.Errorf("Expected session token 'session-token', got %s", cfg.SessionTokens["alice"])
	}

	// Test GetAuthHeader
	header, err := GetAuthHeader()
	if err != nil {
		t.Fatalf("GetAuthHeader failed: %v", err)
	}
	if header != "Bearer session-token" {
		t.Errorf("Expected Bearer token, got %s", header)
	}
}
