package server

import (
	"context"
	"fmt"
	"log"
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
		log.Println("No .env file found, using defaults/env vars")
	}

	// Initialize BlobStore
	var blobStore BlobStore
	storageType := os.Getenv("STORAGE_TYPE")

	if storageType == "s3" {
		bucket := os.Getenv("AWS_BUCKET")
		if bucket == "" {
			return nil, fmt.Errorf("AWS_BUCKET required for s3 storage")
		}
		log.Printf("Using S3 Storage (Bucket: %s)", bucket)
		var err error
		blobStore, err = NewS3BlobStore(context.Background(), bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to init S3 storage: %w", err)
		}
	} else {
		log.Printf("Using Local Storage (Dir: %s)", storageDir)
		blobStore = NewLocalBlobStore(storageDir)
	}

	store, err := NewStorage(storageDir, blobStore)
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
	log.Printf("Server listening on %s", s.Port)
	return s.Server.ListenAndServe()
}
