package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("token", "", "Registration token (if required by server)")
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the current user with the server",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")

		if cfg.CurrentUsername == "" {
			fmt.Println("No current user set. Use 'config init' first.")
			return
		}

		// Get Keys
		if _, ok := cfg.IdentityPrivateKeys[cfg.CurrentUsername]; !ok {
			fmt.Println("Identity private key not found.")
			return
		}
		if _, ok := cfg.ExchangePrivateKeys[cfg.CurrentUsername]; !ok {
			fmt.Println("Exchange private key not found.")
			return
		}

		// Re-derive public keys (or store them in config? We store them in Users map, but let's be safe)
		// Actually, we can just use the ones in cfg.Users[username]
		user, ok := cfg.Users[cfg.CurrentUsername]
		if !ok {
			fmt.Println("User not found in config.")
			return
		}

		// If keys are missing in User struct (e.g. old config), re-derive?
		// For now, assume they are there. But wait, we store private keys.
		// We can re-derive public keys from private keys if needed, but Ed25519 private key includes public key.
		// X25519 private key does NOT include public key easily without re-generating?
		// Actually, curve25519.ScalarBaseMult does it.
		// But let's trust cfg.Users for now.

		fmt.Printf("Registering user %s...\n", cfg.CurrentUsername)

		data, err := json.Marshal(user)
		if err != nil {
			fmt.Println("Error marshaling user:", err)
			return
		}

		req, err := http.NewRequest("POST", cfg.ServerURL+"/users", bytes.NewBuffer(data))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("X-Registration-Token", token)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error registering user:", err)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Registration failed: %s\n", string(body))
			return
		}

		fmt.Println("User registered successfully!")
	},
}
