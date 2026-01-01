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
			// Try to fetch from server? For now, just error.
			fmt.Printf("Unknown user: %s. Add them with 'add-user' first.\n", recipient)
			return
		}
		var recipientPub [32]byte
		copy(recipientPub[:], recipientUser.PublicKey)

		// Get Sender Private Key
		senderPrivBytes, ok := cfg.PrivateKeys[cfg.CurrentUsername]
		if !ok {
			fmt.Printf("No private key for current user %s\n", cfg.CurrentUsername)
			return
		}
		var senderPriv [32]byte
		copy(senderPriv[:], senderPrivBytes)

		// Read File
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// Encrypt File

		// 2. Encrypt Content with Symmetric Key (using Box for simplicity? No, Box is asymmetric.
		// Wait, my crypto package only exposes Box (Asymmetric).
		// I need a Symmetric encryption function (e.g. SecretBox).
		// I missed this in the crypto package implementation.
		// I will use Box to encrypt the file content directly for now if file is small?
		// No, that's bad practice. I should add SecretBox to crypto package.
		// OR, since I'm already editing this file, I can just use Box for everything if I treat the symmetric key as a "shared key" but Box doesn't work like that.
		// I'll update the crypto package to include SecretBox or just use Box for the file content using a generated ephemeral keypair?
		// Actually, the standard way is:
		// File Key (Symmetric) -> Encrypt File
		// File Key -> Encrypt with Recipient PubKey (Asymmetric)

		// Let's pause and update crypto package to include SecretBox (NaCl Symmetric).
		// I'll assume I have `crypto.EncryptSymmetric` and `crypto.DecryptSymmetric`.
		// I will implement them in the next step or right now via a separate tool call?
		// I can't do it in the middle of this file.
		// I'll use a placeholder or just use Box for the file content using a temporary keypair for the file itself?
		// That's effectively what Box does (ephemeral sender).
		// So:
		// 1. Generate Ephemeral KeyPair (E_pub, E_priv)
		// 2. Encrypt File with (E_priv, Recipient_pub) -> Ciphertext
		// 3. Send (Ciphertext, E_pub)
		// This works! No need for symmetric key wrapping if we rely on Box's ephemeral nature.
		// But wait, if I send to myself, I need my private key to decrypt.
		// If I use Box, I encrypt with (My_priv, Recipient_pub).
		// Recipient decrypts with (Recipient_priv, My_pub).
		// This requires the recipient to know "My_pub".
		// If I use an ephemeral sender, the recipient needs to know "E_pub".
		// So I can just attach "E_pub" to the message.
		// This is standard "Anonymous Sender" Box.

		// So:
		// 1. Generate Ephemeral KeyPair (Ephemeral).
		// 2. Encrypt File with (Ephemeral.Private, Recipient.Public).
		// 3. Send (EncryptedContent, Ephemeral.Public).
		// The "EncryptedKey" field in metadata can store the Ephemeral Public Key.
		// This avoids needing a separate symmetric cipher if the file fits in memory (Box is fine for reasonable sizes, but stream is better for large files. For MVP, memory is fine).

		fmt.Println("Encrypting file...")
		ephemeral, err := crypto.GenerateKeyPair()
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
				Sender:       cfg.CurrentUsername, // Claimed sender
				Recipient:    recipient,
				FileName:     filepath.Base(filePath),
				EncryptedKey: ephemeral.Public[:], // Store ephemeral public key here
				AutoDelete:   autoDelete,
			},
			EncryptedContent: encryptedContent,
		}

		data, err := json.Marshal(req)
		if err != nil {
			fmt.Println("Error marshaling request:", err)
			return
		}

		resp, err := http.Post(cfg.ServerURL+"/files", "application/json", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error uploading file:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Upload failed: %s\n", string(body))
			return
		}

		fmt.Println("File sent successfully!")
	},
}
