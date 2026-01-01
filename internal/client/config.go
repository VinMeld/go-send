package client

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/VinMeld/go-send/internal/models"
	"github.com/VinMeld/go-send/internal/transport"
)

type Config struct {
	CurrentUsername     string                 `json:"current_username"`
	Users               map[string]models.User `json:"users"`                 // Known users (address book)
	IdentityPrivateKeys map[string][]byte      `json:"identity_private_keys"` // Map username -> Ed25519 private key
	ExchangePrivateKeys map[string][]byte      `json:"exchange_private_keys"` // Map username -> X25519 private key
	SessionTokens       map[string]string      `json:"session_tokens"`        // Map username -> session token
	ServerURL           string                 `json:"server_url"`
	LastListedFiles     []string               `json:"last_listed_files,omitempty"` // Cache for index-based access
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Users:               make(map[string]models.User),
				IdentityPrivateKeys: make(map[string][]byte),
				ExchangePrivateKeys: make(map[string][]byte),
				SessionTokens:       make(map[string]string),
				ServerURL:           transport.DefaultServerURL,
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Users == nil {
		cfg.Users = make(map[string]models.User)
	}
	if cfg.IdentityPrivateKeys == nil {
		cfg.IdentityPrivateKeys = make(map[string][]byte)
	}
	if cfg.ExchangePrivateKeys == nil {
		cfg.ExchangePrivateKeys = make(map[string][]byte)
	}
	if cfg.SessionTokens == nil {
		cfg.SessionTokens = make(map[string]string)
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "go-send", "config.json"), nil
}
