package server

import (
	"log/slog"
	"net/http"
)

// DeleteFile deletes a file if the user is the sender or recipient.
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	// Get authenticated user
	username, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	meta, ok := h.Storage.GetFileMetadata(r.Context(), id)
	if !ok {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Check permission
	if meta.Sender != username && meta.Recipient != username {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.Storage.DeleteFile(r.Context(), id); err != nil {
		slog.Error("failed to delete file", "id", id, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("file deleted", "id", id, "by", username)
	w.WriteHeader(http.StatusOK)
}
