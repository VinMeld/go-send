package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/VinMeld/go-send/internal/models"
	"github.com/google/uuid"
)

type Handler struct {
	Storage *Storage
}

func NewHandler(storage *Storage) *Handler {
	return &Handler{Storage: storage}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Username == "" || len(user.PublicKey) == 0 {
		http.Error(w, "invalid user", http.StatusBadRequest)
		return
	}

	if err := h.Storage.AddUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		// Try path param if using router, but for stdlib:
		// We'll stick to query param or simple path parsing if needed.
		// Let's assume /users?username=... for simplicity or parse path.
		// Actually, let's use a simple mux in main.go, so here we might just expect query param.
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	user, ok := h.Storage.GetUser(username)
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	var req models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	if err := h.Storage.SaveFile(req.Metadata, req.EncryptedContent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req.Metadata)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	recipient := r.URL.Query().Get("recipient")
	if recipient == "" {
		http.Error(w, "recipient required", http.StatusBadRequest)
		return
	}

	files := h.Storage.ListFiles(recipient)
	json.NewEncoder(w).Encode(files)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	// Simple path parsing: /files/{id}
	// But since we are using stdlib http.ServeMux (likely), we might need to parse URL.
	// Let's assume query param ?id=... for simplicity unless we use a router.
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	meta, ok := h.Storage.GetFileMetadata(id)
	if !ok {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	content, err := h.Storage.GetFileContent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.UploadRequest{
		Metadata:         meta,
		EncryptedContent: content,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// If encoding fails, we probably shouldn't delete the file yet?
		// Or maybe we should log it.
		return
	}

	// Auto-delete after successful download if requested
	if meta.AutoDelete {
		h.Storage.DeleteFile(id)
	}
}
