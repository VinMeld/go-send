package main

import (
	"fmt"
	"os"

	"github.com/VinMeld/go-send/internal/client"
)

func main() {
	if err := client.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
