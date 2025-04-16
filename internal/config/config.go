package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey       string
	TokenExpireTime time.Duration
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
}

func (c *Config) Validate() error {
	if c.OAuth.GoogleClientID == "" {
		return fmt.Errorf("missing Google client ID")
	}
	if c.OAuth.GoogleClientSecret == "" {
		return fmt.Errorf("missing Google client secret")
	}
	if c.OAuth.RedirectURL == "" {
		return fmt.Errorf("missing OAuth redirect URL")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AddConfigPath("./config")

	//Enable env
	viper.AutomaticEnv()

	// Map env variable to config above
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("jwt.secret", "JWT_SECRET")

	// Map oauth env variables
	viper.BindEnv("oauth.googleClientID", "GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.googleClientSecret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.redirectURL", "OAUTH_REDIRECT_URL")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &config, nil
}
