package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadFileCmd)
}

var downloadFileCmd = &cobra.Command{
	Use:   "download-file <file_id>",
	Short: "Download and decrypt a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileID := args[0]
		if cfg.CurrentUsername == "" {
			fmt.Println("No current user set. Use 'set-user'.")
			return
		}

		// Fetch File
		resp, err := http.Get(fmt.Sprintf("%s/files/download?id=%s", cfg.ServerURL, fileID))
		if err != nil {
			fmt.Println("Error fetching file:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Server returned error:", resp.Status)
			return
		}

		var req models.UploadRequest // Reusing this struct as response wrapper
		if err := json.NewDecoder(resp.Body).Decode(&req); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		// Decrypt
		// 1. Get Recipient Private Key
		privKeyBytes, ok := cfg.PrivateKeys[cfg.CurrentUsername]
		if !ok {
			fmt.Println("Private key not found for current user")
			return
		}
		var recipientPriv [32]byte
		copy(recipientPriv[:], privKeyBytes)

		// 2. Get Sender Public Key (Ephemeral) from Metadata
		var senderPub [32]byte
		if len(req.Metadata.EncryptedKey) != 32 {
			fmt.Println("Invalid ephemeral public key length in metadata")
			return
		}
		copy(senderPub[:], req.Metadata.EncryptedKey)

		// 3. Decrypt Content
		decrypted, err := crypto.Decrypt(req.EncryptedContent, &senderPub, &recipientPriv)
		if err != nil {
			fmt.Println("Error decrypting file:", err)
			return
		}

		// Save to Disk
		outputFile := req.Metadata.FileName
		// Prevent overwriting? For now, just write.
		if err := os.WriteFile(outputFile, decrypted, 0644); err != nil {
			fmt.Println("Error saving file:", err)
			return
		}

		fmt.Printf("File downloaded and decrypted to %s\n", outputFile)
	},
}
