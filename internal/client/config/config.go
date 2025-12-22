package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ClientConfig defines configuration parameters for the GophKeeper client.
// Values may be loaded from a client.yaml file or fall back to defaults.
type ClientConfig struct {
	ServerAddress string `mapstructure:"server_address"`
	WorkDir       string `mapstructure:"work_dir"`
	UseTLS        bool   `mapstructure:"use_tls"`
	Version       string `mapstructure:"version"`
	BuildTime     string `mapstructure:"build_time"`
}

// Load loads the client configuration from a client.yaml file located in the
// current working directory. If the file does not exist, default values are
// applied. The version and buildTime parameters override corresponding fields.
// Returns the populated ClientConfig or an error if unmarshalling fails.
func Load(version, buildTime string) (*ClientConfig, error) {
	viper.SetConfigName("client")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("server_address", "http://localhost:8080")

	home, _ := os.UserHomeDir()
	viper.SetDefault("work_dir", filepath.Join(home, ".passkeeper"))
	viper.SetDefault("use_tls", false)
	viper.SetDefault("version", version)
	viper.SetDefault("build_time", buildTime)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var cfg ClientConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshal config: %w", err)
	}

	return &cfg, nil
}
