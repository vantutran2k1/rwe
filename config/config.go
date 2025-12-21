package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	GRPCPort string `mapstructure:"grpc_port"`
	HTTPPort string `mapstructure:"http_port"`
}

type AuthConfig struct {
	TokenSymmetricKey  string `mapstructure:"token_symmetric_key"`
	TokenDurationHours int32  `mapstructure:"token_duration_hours"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

func Load(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
