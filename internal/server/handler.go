package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/VinMeld/go-send/internal/models"
	"github.com/google/uuid"
)

type contextKey string

const (
	userContextKey contextKey = "user"
)

type Handler struct {
	Storage           *Storage
	RegistrationToken string
}

func NewHandler(storage *Storage) *Handler {
	return &Handler{Storage: storage}
}

// SetRegistrationToken sets the registration token for the handler.
func (h *Handler) SetRegistrationToken(token string) {
	h.RegistrationToken = token
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if h.RegistrationToken != "" {
		token := r.Header.Get("X-Registration-Token")
		if token != h.RegistrationToken {
			slog.Warn("invalid registration token", "token", token)
			http.Error(w, "forbidden: invalid registration token", http.StatusForbidden)
			return
		}
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.Error("failed to decode user", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Username == "" || len(user.IdentityPublicKey) == 0 || len(user.ExchangePublicKey) == 0 {
		http.Error(w, "invalid user", http.StatusBadRequest)
		return
	}

	if err := h.Storage.AddUser(r.Context(), user); err != nil {
		slog.Error("failed to add user", "username", user.Username, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("user registered", "username", user.Username)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		// List all users
		users, err := h.Storage.ListAllUsers(r.Context())
		if err != nil {
			slog.Error("failed to list users", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(users)
		return
	}

	user, ok := h.Storage.GetUser(r.Context(), username)
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode upload request", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate
	if req.Metadata.Recipient == "" || len(req.EncryptedContent) == 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Assign ID and Timestamp
	req.Metadata.ID = uuid.New().String()
	req.Metadata.Timestamp = time.Now()

	if err := h.Storage.SaveFile(r.Context(), req.Metadata, req.EncryptedContent); err != nil {
		slog.Error("failed to save file", "sender", req.Metadata.Sender, "recipient", req.Metadata.Recipient, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("file uploaded", "id", req.Metadata.ID, "sender", req.Metadata.Sender, "recipient", req.Metadata.Recipient)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req.Metadata)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	recipient := r.URL.Query().Get("recipient")
	if recipient == "" {
		http.Error(w, "recipient required", http.StatusBadRequest)
		return
	}

	files, err := h.Storage.ListFiles(r.Context(), recipient)
	if err != nil {
		slog.Error("failed to list files", "recipient", recipient, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(files)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	meta, ok := h.Storage.GetFileMetadata(r.Context(), id)
	if !ok {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	content, err := h.Storage.GetFileContent(id)
	if err != nil {
		slog.Error("failed to get file content", "id", id, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("file downloaded", "id", id, "recipient", meta.Recipient)

	resp := models.UploadRequest{
		Metadata:         meta,
		EncryptedContent: content,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}

	// Auto-delete after successful download if requested
	if meta.AutoDelete {
		_ = h.Storage.DeleteFile(r.Context(), id)
	}
}

// AuthMiddleware protects routes by requiring a valid session token.
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		session, ok := h.Storage.GetSession(r.Context(), token)
		if !ok || session.ExpiresAt.Before(time.Now()) {
			if ok {
				_ = h.Storage.DeleteSession(r.Context(), token)
			}
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, session.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simple router
	switch r.URL.Path {
	case "/ping":
		h.Ping(w, r)
	case "/users":
		if r.Method == http.MethodPost {
			h.RegisterUser(w, r)
		} else if r.Method == http.MethodGet {
			h.GetUser(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/files":
		if r.Method == http.MethodPost {
			h.AuthMiddleware(h.UploadFile)(w, r)
		} else if r.Method == http.MethodGet {
			h.AuthMiddleware(h.ListFiles)(w, r)
		} else if r.Method == http.MethodDelete {
			h.AuthMiddleware(h.DeleteFile)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/files/download":
		if r.Method == http.MethodGet {
			h.AuthMiddleware(h.DownloadFile)(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/auth/challenge":
		if r.Method == http.MethodGet {
			h.HandleGetChallenge(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/auth/login":
		if r.Method == http.MethodPost {
			h.HandleLogin(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}
