package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Login(); err != nil {
			fmt.Println("Login failed:", err)
			return
		}
		fmt.Println("Logged in successfully!")
	},
}

// Login performs the challenge-response authentication flow.
func Login() error {
	if cfg.CurrentUsername == "" {
		return fmt.Errorf("no current user set")
	}

	// 1. Get Challenge
	resp, err := http.Get(fmt.Sprintf("%s/auth/challenge?username=%s", cfg.ServerURL, cfg.CurrentUsername))
	if err != nil {
		return fmt.Errorf("failed to get challenge: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s", string(body))
	}

	var challenge models.AuthChallenge
	if err := json.NewDecoder(resp.Body).Decode(&challenge); err != nil {
		return fmt.Errorf("failed to decode challenge: %w", err)
	}

	// 2. Sign Challenge
	privKey, ok := cfg.IdentityPrivateKeys[cfg.CurrentUsername]
	if !ok {
		return fmt.Errorf("identity private key not found for user %s", cfg.CurrentUsername)
	}

	signature := crypto.Sign(privKey, []byte(challenge.Nonce))

	// 3. Send Response
	authResp := models.AuthResponse{
		Username:  cfg.CurrentUsername,
		Nonce:     challenge.Nonce,
		Signature: signature,
	}
	data, _ := json.Marshal(authResp)

	resp, err = http.Post(cfg.ServerURL+"/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send login response: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(body))
	}

	var session models.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return fmt.Errorf("failed to decode session: %w", err)
	}

	// 4. Store Token
	cfg.SessionTokens[cfg.CurrentUsername] = session.Token
	return SaveConfigGlobal()
}

// GetAuthHeader returns the Authorization header value for the current user.
func GetAuthHeader() (string, error) {
	token, ok := cfg.SessionTokens[cfg.CurrentUsername]
	if !ok {
		// Try to login automatically?
		if err := Login(); err != nil {
			return "", fmt.Errorf("not logged in and automatic login failed: %w", err)
		}
		token = cfg.SessionTokens[cfg.CurrentUsername]
	}
	return "Bearer " + token, nil
}
