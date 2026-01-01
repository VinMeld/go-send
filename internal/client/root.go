package client

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *Config
)

var rootCmd = &cobra.Command{
	Use:   "go-send",
	Short: "Secure file sending CLI",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/go-send/config.json)")
}

func initConfig() {
	var err error
	path := cfgFile
	if path == "" {
		path, err = GetConfigPath()
		if err != nil {
			fmt.Println("Error getting config path:", err)
			os.Exit(1)
		}
	}

	cfg, err = LoadConfig(path)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func GetConfig() *Config {
	return cfg
}

func SaveConfigGlobal() error {
	path := cfgFile
	if path == "" {
		var err error
		path, err = GetConfigPath()
		if err != nil {
			return err
		}
	}
	return SaveConfig(path, cfg)
}
