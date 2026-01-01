package client

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteFileCmd)
}

var deleteFileCmd = &cobra.Command{
	Use:   "delete-file <file_id>",
	Short: "Delete a file from the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]

		if cfg.CurrentUsername == "" {
			fmt.Println("No current user set. Use 'config init' first.")
			return
		}

		authHeader, err := GetAuthHeader()
		if err != nil {
			fmt.Println("Authentication error:", err)
			return
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files?id=%s", cfg.ServerURL, fileID), nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error deleting file:", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Server returned error:", resp.Status)
			return
		}

		fmt.Println("File deleted successfully!")
	},
}
