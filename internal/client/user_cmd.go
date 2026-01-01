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

		// Generate Identity Keys (Ed25519)
		idKeys, err := crypto.GenerateIdentityKeyPair()
		if err != nil {
			fmt.Println("Error generating identity keys:", err)
			return
		}

		// Generate Exchange Keys (X25519)
		exKeys, err := crypto.GenerateExchangeKeyPair()
		if err != nil {
			fmt.Println("Error generating exchange keys:", err)
			return
		}

		// Update Config
		cfg.CurrentUsername = username
		cfg.IdentityPrivateKeys[username] = idKeys.Private
		cfg.ExchangePrivateKeys[username] = exKeys.Private[:]
		cfg.Users[username] = models.User{
			Username:          username,
			IdentityPublicKey: idKeys.Public,
			ExchangePublicKey: exKeys.Public[:],
		}

		if err := SaveConfigGlobal(); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}
		fmt.Printf("Initialized user %s\n", username)
		fmt.Printf("Identity Public Key: %s\n", base64.StdEncoding.EncodeToString(idKeys.Public))
		fmt.Printf("Exchange Public Key: %s\n", base64.StdEncoding.EncodeToString(exKeys.Public[:]))
	},
}

var setUserCmd = &cobra.Command{
	Use:   "set-user <username>",
	Short: "Set current active user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		if _, ok := cfg.IdentityPrivateKeys[username]; !ok {
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
	Use:   "add-user <username> <id_pub_key_b64> <ex_pub_key_b64>",
	Short: "Add a known user",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		idPubKeyStr := args[1]
		exPubKeyStr := args[2]

		idPubKey, err := base64.StdEncoding.DecodeString(idPubKeyStr)
		if err != nil {
			fmt.Println("Error decoding identity public key:", err)
			return
		}
		exPubKey, err := base64.StdEncoding.DecodeString(exPubKeyStr)
		if err != nil {
			fmt.Println("Error decoding exchange public key:", err)
			return
		}

		cfg.Users[username] = models.User{
			Username:          username,
			IdentityPublicKey: idPubKey,
			ExchangePublicKey: exPubKey,
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
			fmt.Printf("- %s\n", u.Username)
			fmt.Printf("  Identity: %s\n", base64.StdEncoding.EncodeToString(u.IdentityPublicKey))
			fmt.Printf("  Exchange: %s\n", base64.StdEncoding.EncodeToString(u.ExchangePublicKey))
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
