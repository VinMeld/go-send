package main

import (
	"flag"
	"log"
	"os"

	"github.com/VinMeld/go-send/internal/server"
	"github.com/VinMeld/go-send/internal/transport"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (overrides env PORT)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = transport.DefaultServerPort
	}

	storageDir := "./server_data"
	if len(flag.Args()) > 0 {
		storageDir = flag.Args()[0]
	}

	srv, err := server.NewServer(port, storageDir)
	if err != nil {
		log.Fatalf("Failed to init server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
