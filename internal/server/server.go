package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Server represents the HTTP server.
type Server struct {
	Port    string
	Storage *Storage
	Handler *Handler
	Server  *http.Server
}

// NewServer initializes a new Server.
func NewServer(port string, storageDir string) (*Server, error) {
	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using defaults/env vars")
	}

	storageType := os.Getenv("STORAGE_TYPE")
	var store *Storage
	var err error

	if storageType == "s3" {
		bucket := os.Getenv("AWS_BUCKET")
		if bucket == "" {
			return nil, fmt.Errorf("AWS_BUCKET required for s3 storage")
		}
		slog.Info("Using S3 Storage", "bucket", bucket)
		region := os.Getenv("AWS_REGION")
		blobStore, err := NewS3BlobStore(context.Background(), bucket, region)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 blob store: %w", err)
		}
		store, err = NewStorage(storageDir, blobStore)
	} else {
		if storageDir == "" {
			storageDir = os.Getenv("DATA_DIR")
		}
		if storageDir == "" {
			storageDir = "server_data"
		}
		slog.Info("Using Local Storage", "dir", storageDir)
		blobStore := NewLocalBlobStore(storageDir)
		store, err = NewStorage(storageDir, blobStore)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	h := NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Ping(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.RegisterUser(w, r)
		} else if r.Method == http.MethodGet {
			h.GetUser(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.UploadFile(w, r)
		} else if r.Method == http.MethodGet {
			h.ListFiles(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/files/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.DownloadFile(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Ensure port has colon
	if port == "" {
		port = ":8080"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	return &Server{
		Port:    port,
		Storage: store,
		Handler: h,
		Server:  &http.Server{Addr: port, Handler: mux},
	}, nil
}

// Start starts the server.
func (s *Server) Start() error {
	slog.Info("Server starting", "addr", s.Server.Addr)
	return s.Server.ListenAndServe()
}
