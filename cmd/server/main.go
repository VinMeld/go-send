package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/VinMeld/go-send/internal/server"
	"github.com/VinMeld/go-send/internal/transport"
)

func main() {
	var port string
	flag.StringVar(&port, "port", transport.DefaultServerPort, "Server port")
	flag.Parse()

	// Handle storage dir which might be a positional arg after flags?
	// The original code used os.Args[1]. flag.Parse() consumes flags.
	// Remaining args are in flag.Args().
	storageDir := "./server_data"
	if len(flag.Args()) > 0 {
		storageDir = flag.Args()[0]
	}

	store, err := server.NewStorage(storageDir)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	h := server.NewHandler(store)

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
	if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
