package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
	"github.com/google/uuid"
)

// HandleGetChallenge generates a random nonce for the user.
func (h *Handler) HandleGetChallenge(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, ok := h.Storage.GetUser(r.Context(), username); !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Generate random nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		http.Error(w, "failed to generate nonce", http.StatusInternalServerError)
		return
	}
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)

	if err := h.Storage.CreateChallenge(r.Context(), username, nonce); err != nil {
		slog.Error("failed to create challenge", "username", username, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("challenge created", "username", username)
	_ = json.NewEncoder(w).Encode(models.AuthChallenge{
		Username: username,
		Nonce:    nonce,
	})
}

// HandleLogin verifies the challenge response and issues a session token.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var resp models.AuthResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Retrieve challenge
	expectedNonce, ok := h.Storage.GetChallenge(r.Context(), resp.Username)
	if !ok || expectedNonce != resp.Nonce {
		http.Error(w, "invalid or expired challenge", http.StatusUnauthorized)
		return
	}

	// Get user's public identity key
	user, ok := h.Storage.GetUser(r.Context(), resp.Username)
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Verify signature
	if !crypto.Verify(user.IdentityPublicKey, []byte(resp.Nonce), resp.Signature) {
		slog.Warn("invalid login signature", "username", resp.Username)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Create session
	token := uuid.New().String()
	session := models.Session{
		Token:     token,
		Username:  resp.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := h.Storage.CreateSession(r.Context(), session); err != nil {
		slog.Error("failed to create session", "username", resp.Username, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("user logged in", "username", resp.Username)
	_ = json.NewEncoder(w).Encode(session)
}
