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

		resp, err := http.Get(fmt.Sprintf("%s/files?recipient=%s", cfg.ServerURL, cfg.CurrentUsername))
		if err != nil {
			fmt.Println("Error fetching files:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Server returned error:", resp.Status)
			return
		}

		var files []models.FileMetadata
		if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		fmt.Printf("Files for %s:\n", cfg.CurrentUsername)
		for _, f := range files {
			fmt.Printf("- [%s] %s (from %s) - %s\n", f.ID, f.FileName, f.Sender, f.Timestamp.Format(time.RFC822))
		}
	},
}
