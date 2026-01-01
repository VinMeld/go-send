package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check connection to the server",
	Run: func(cmd *cobra.Command, args []string) {
		url := cfg.ServerURL
		if url == "" {
			fmt.Println("Server URL not set in config")
			return
		}

		fmt.Printf("Pinging %s...\n", url)
		start := time.Now()
		resp, err := http.Get(url + "/ping")
		if err != nil {
			fmt.Printf("Failed to ping server: %v\n", err)
			return
		}
		defer resp.Body.Close()
		duration := time.Since(start)

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Pong! Server is reachable (Latency: %v)\n", duration)
		} else {
			fmt.Printf("Server returned status: %s\n", resp.Status)
		}
	},
}
