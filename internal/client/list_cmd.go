package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VinMeld/go-send/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listFilesCmd)
}

var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List files waiting for the current user",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.CurrentUsername == "" {
			fmt.Println("No current user set. Use 'set-user'.")
			return
		}

		authHeader, err := GetAuthHeader()
		if err != nil {
			fmt.Println("Authentication error:", err)
			return
		}

		httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/files?recipient=%s", cfg.ServerURL, cfg.CurrentUsername), nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		httpReq.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Println("Error fetching files:", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Server returned error:", resp.Status)
			return
		}

		var files []models.FileMetadata
		if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		// Update cache
		cfg.LastListedFiles = make([]string, 0, len(files))
		for _, f := range files {
			cfg.LastListedFiles = append(cfg.LastListedFiles, f.ID)
		}
		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Warning: Failed to save file list cache:", err)
		}

		fmt.Printf("Files for %s:\n", cfg.CurrentUsername)
		for i, f := range files {
			fmt.Printf("%d - [%s] %s (from %s) - %s\n", i+1, f.ID, f.FileName, f.Sender, f.Timestamp.Format(time.RFC822))
		}
	},
}
