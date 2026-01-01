package client

import (
	"encoding/base64"
	"fmt"

	"github.com/VinMeld/go-send/internal/crypto"
	"github.com/VinMeld/go-send/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(setUserCmd)
	rootCmd.AddCommand(setServerCmd)
	rootCmd.AddCommand(addUserCmd)
	rootCmd.AddCommand(listUsersCmd)
	rootCmd.AddCommand(removeUserCmd)
}

var setServerCmd = &cobra.Command{
	Use:   "set-server <url>",
	Short: "Set the remote server URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		cfg.ServerURL = url
		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Server URL set to %s\n", url)
	},
}

var configInitCmd = &cobra.Command{
	Use:   "config init --user <username>",
	Short: "Initialize configuration and generate keys",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("user")
		if username == "" {
			fmt.Println("Username required")
			return
		}

		// Generate Keys
		keys, err := crypto.GenerateKeyPair()
		if err != nil {
			fmt.Println("Error generating keys:", err)
			return
		}

		// Update Config
		cfg.CurrentUsername = username
		cfg.PrivateKeys[username] = keys.Private[:]
		cfg.Users[username] = models.User{
			Username:  username,
			PublicKey: keys.Public[:],
		}

		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Initialized user %s\n", username)
		fmt.Printf("Public Key: %s\n", base64.StdEncoding.EncodeToString(keys.Public[:]))
	},
}

var setUserCmd = &cobra.Command{
	Use:   "set-user <username>",
	Short: "Set current active user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		if _, ok := cfg.PrivateKeys[username]; !ok {
			fmt.Printf("User %s not found in local config (no private key)\n", username)
			return
		}
		cfg.CurrentUsername = username
		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Current user set to %s\n", username)
	},
}

var addUserCmd = &cobra.Command{
	Use:   "add-user <username> <public_key_base64>",
	Short: "Add a known user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		pubKeyStr := args[1]

		pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
		if err != nil {
			fmt.Println("Error decoding public key:", err)
			return
		}
		if len(pubKey) != 32 {
			fmt.Println("Invalid public key length (must be 32 bytes)")
			return
		}

		cfg.Users[username] = models.User{
			Username:  username,
			PublicKey: pubKey,
		}
		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Added user %s\n", username)
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "List known users",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Known Users:")
		for _, u := range cfg.Users {
			fmt.Printf("- %s (Public Key: %s)\n", u.Username, base64.StdEncoding.EncodeToString(u.PublicKey))
		}
	},
}

var removeUserCmd = &cobra.Command{
	Use:   "remove-user <username>",
	Short: "Remove a known user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		delete(cfg.Users, username)
		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Removed user %s\n", username)
	},
}

func init() {
	configInitCmd.Flags().String("user", "", "Username to initialize")
}
