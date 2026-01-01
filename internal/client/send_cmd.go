package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sendFileCmd)
	sendFileCmd.Flags().Bool("auto-delete", false, "Delete file from server after download")
}

var sendFileCmd = &cobra.Command{
	Use:   "send-file [recipient] <file>",
	Short: "Send an encrypted file",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		autoDelete, _ := cmd.Flags().GetBool("auto-delete")

		var recipient string
		var filePath string

		if len(args) == 1 {
			// send-file <file> -> Recipient is self
			recipient = cfg.CurrentUsername
			filePath = args[0]
		} else {
			// send-file <recipient> <file>
			recipient = args[0]
			filePath = args[1]
		}

		if recipient == "" {
			fmt.Println("Recipient not specified and no current user set")
			return
		}

		// Get Recipient Public Key
		recipientUser, ok := cfg.Users[recipient]
		if !ok {
			fmt.Printf("User '%s' not found locally. Searching on server...\n", recipient)
			resp, err := http.Get(cfg.ServerURL + "/users?username=" + recipient)
			if err != nil {
				fmt.Printf("Error contacting server: %v\n", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == http.StatusOK {
				var foundUser models.User
				if err := json.NewDecoder(resp.Body).Decode(&foundUser); err != nil {
					fmt.Printf("Error decoding user from server: %v\n", err)
					return
				}
				if len(foundUser.IdentityPublicKey) == 0 || len(foundUser.ExchangePublicKey) == 0 {
					fmt.Println("Server returned invalid user keys.")
					return
				}

				cfg.Users[recipient] = foundUser
				if err := SaveConfigGlobal(); err != nil {
					fmt.Printf("Warning: Failed to save user to local config: %v\n", err)
				} else {
					fmt.Printf("Found user '%s' and added to address book.\n", recipient)
				}
				recipientUser = foundUser
			} else {
				fmt.Printf("Unknown user: %s. Add them with 'add-user' first or ensure they are registered.\n", recipient)
				return
			}
		}
		var recipientPub [32]byte
		copy(recipientPub[:], recipientUser.ExchangePublicKey)

		// Read File
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		fmt.Println("Encrypting file...")
		ephemeral, err := crypto.GenerateExchangeKeyPair()
		if err != nil {
			fmt.Println("Error generating ephemeral key:", err)
			return
		}

		encryptedContent, err := crypto.Encrypt(fileContent, &recipientPub, ephemeral.Private)
		if err != nil {
			fmt.Println("Error encrypting file:", err)
			return
		}

		// Upload
		req := models.UploadRequest{
			Metadata: models.FileMetadata{
				Sender:       cfg.CurrentUsername,
				Recipient:    recipient,
				FileName:     filepath.Base(filePath),
				EncryptedKey: ephemeral.Public[:],
				AutoDelete:   autoDelete,
			},
			EncryptedContent: encryptedContent,
		}

		data, err := json.Marshal(req)
		if err != nil {
			fmt.Println("Error marshaling request:", err)
			return
		}

		authHeader, err := GetAuthHeader()
		if err != nil {
			fmt.Println("Authentication error:", err)
			return
		}

		reqBody, err := http.NewRequest("POST", cfg.ServerURL+"/files", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		reqBody.Header.Set("Content-Type", "application/json")
		reqBody.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(reqBody)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Upload failed: %s\n", string(body))
			return
		}

		fmt.Println("File sent successfully!")
	},
}
