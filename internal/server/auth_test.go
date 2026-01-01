package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
)

func TestAuthHandlers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-auth-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	storage, err := NewStorage(tmpDir, NewLocalBlobStore(tmpDir))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(storage)

	// Setup user
	idKey, _ := crypto.GenerateIdentityKeyPair()
	exKey, _ := crypto.GenerateExchangeKeyPair()
	user := models.User{
		Username:          "alice",
		IdentityPublicKey: idKey.Public,
		ExchangePublicKey: exKey.Public[:],
	}
	if err := storage.AddUser(user); err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Test HandleGetChallenge
	req, _ := http.NewRequest("GET", "/auth/challenge?username=alice", nil)
	rr := httptest.NewRecorder()
	handler.HandleGetChallenge(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleGetChallenge returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var challenge models.AuthChallenge
	_ = json.NewDecoder(rr.Body).Decode(&challenge)
	if challenge.Nonce == "" {
		t.Error("Expected nonce in challenge")
	}

	// Test HandleLogin
	signature := crypto.Sign(idKey.Private, []byte(challenge.Nonce))
	authResp := models.AuthResponse{
		Username:  "alice",
		Nonce:     challenge.Nonce,
		Signature: signature,
	}
	body, _ := json.Marshal(authResp)

	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	handler.HandleLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleLogin returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var session models.Session
	_ = json.NewDecoder(rr.Body).Decode(&session)
	if session.Token == "" {
		t.Error("Expected session token")
	}

	// Verify session in storage
	storedSession, ok := storage.GetSession(session.Token)
	if !ok {
		t.Error("Session not stored")
	}
	if storedSession.Username != "alice" {
		t.Errorf("Expected session username alice, got %s", storedSession.Username)
	}
}

func TestAuthErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-auth-error-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	storage, err := NewStorage(tmpDir, NewLocalBlobStore(tmpDir))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(storage)

	// Test GetChallenge - User Not Found
	req, _ := http.NewRequest("GET", "/auth/challenge?username=unknown", nil)
	rr := httptest.NewRecorder()
	handler.HandleGetChallenge(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for unknown user, got %d", rr.Code)
	}

	// Test Login - Invalid Challenge
	authResp := models.AuthResponse{
		Username:  "alice",
		Nonce:     "invalid",
		Signature: []byte("sig"),
	}
	body, _ := json.Marshal(authResp)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	handler.HandleLogin(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid challenge, got %d", rr.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-send-auth-mw-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	storage, err := NewStorage(tmpDir, NewLocalBlobStore(tmpDir))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(storage)

	// Create session
	session := models.Session{
		Token:     "valid-token",
		Username:  "alice",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	storage.CreateSession(session)

	protectedHandler := handler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextKey).(string)
		if user != "alice" {
			t.Errorf("Expected user alice in context, got %s", user)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Test Valid Token
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	protectedHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid token, got %d", rr.Code)
	}

	// Test Missing Header
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	protectedHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing header, got %d", rr.Code)
	}

	// Test Invalid Token
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rr = httptest.NewRecorder()
	protectedHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", rr.Code)
	}
}
