package client

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/VinMeld/go-send/internal/models"
	"github.com/VinMeld/go-send/internal/transport"
)

type Config struct {
	CurrentUsername string                 `json:"current_username"`
	Users           map[string]models.User `json:"users"`        // Known users (address book)
	PrivateKeys     map[string][]byte      `json:"private_keys"` // Map username -> private key (raw bytes)
	ServerURL       string                 `json:"server_url"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Users:       make(map[string]models.User),
				PrivateKeys: make(map[string][]byte),
				ServerURL:   transport.DefaultServerURL,
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
	if cfg.PrivateKeys == nil {
		cfg.PrivateKeys = make(map[string][]byte)
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
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".go-send", "config.json"), nil
}
