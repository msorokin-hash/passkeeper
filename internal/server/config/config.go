package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the full application configuration. It includes
// server settings, database configuration, and logging parameters.
// Values are typically loaded from a YAML file and environment variables.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig defines settings related to the application server,
// including network address, encryption master key, TLS usage flag,
// and parameters for JWT token generation.
type ServerConfig struct {
	Address       string        `mapstructure:"address"`
	MasterKey     string        `mapstructure:"master_key"`
	UseTLS        bool          `mapstructure:"use_tls"`
	TokenKey      string        `mapstructure:"token_key"`
	TokenLifetime time.Duration `mapstructure:"token_lifetime"`
}

// DatabaseConfig contains connection parameters for the database layer,
// such as host address, port, database name, and user credentials.
// It is used to build DSN strings for specific database drivers.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// LoggingConfig defines settings related to logging output,
// including the desired log level.
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// PGConnectionString constructs and returns the PostgreSQL connection string
// using the database configuration values. SSL mode is disabled by default.
func (c *DatabaseConfig) PGConnectionString() string {
	return "postgresql://" + c.User + ":" + c.Password +
		"@" + c.Host + ":" + c.Port + "/" + c.Name +
		"?sslmode=disable"
}

// Load loads application configuration from the `server.yaml` file,
// falling back to default values if the file is missing. Values may also be
// overridden using environment variables with the prefix `PASSKEEPER_`.
// The resulting configuration is validated before being returned.
func Load() (*Config, error) {
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.token_lifetime", 3600)
	viper.SetDefault("server.use_tls", false)
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("logging.level", "info")

	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PASSKEEPER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("config file not found, using default values")
		} else {
			return nil, fmt.Errorf("error reading configuration: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshal configuration: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validateConfig validates the loaded configuration and ensures that
// required fields are not empty and contain logically valid values.
// It returns an error if any critical configuration parameter is missing.
func validateConfig(config *Config) error {
	if config.Server.Address == "" {
		return fmt.Errorf("server.address required")
	}
	if len(config.Server.MasterKey) < 32 {
		return fmt.Errorf("server.master_key must be at least 32 characters")
	}
	if len(config.Server.TokenKey) < 32 {
		return fmt.Errorf(
			"server.token_key too short (must be >= 32 chars), do not use weak secrets like 'changeit'",
		)
	}
	if config.Server.TokenLifetime <= 0 {
		return fmt.Errorf("server.token_lifetime must be positive")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database.host required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database.name required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database.user required")
	}
	if config.Database.Password == "" {
		return fmt.Errorf("database.password required")
	}

	return nil
}
