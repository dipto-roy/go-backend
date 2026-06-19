package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	JWT      JWTConfig
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Port        string
	Env         string
	LogLevel    string
	PrettyLog   bool
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret         string
	AccessExpiry   time.Duration
	RefreshExpiry  time.Duration
}

type RateLimitConfig struct {
	RequestsPerSecond int64
	Burst             int64
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_PRETTY", true)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	viper.SetDefault("RATE_LIMIT_RPS", 100)
	viper.SetDefault("RATE_LIMIT_BURST", 200)

	secret := viper.GetString("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	dsn := viper.GetString("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}

	connLifetime, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}
	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}
	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	env := strings.ToLower(viper.GetString("APP_ENV"))
	prettyLog := viper.GetBool("LOG_PRETTY")
	if env == "production" {
		prettyLog = false
	}

	return &Config{
		App: AppConfig{
			Port:      viper.GetString("APP_PORT"),
			Env:       env,
			LogLevel:  viper.GetString("LOG_LEVEL"),
			PrettyLog: prettyLog,
		},
		DB: DBConfig{
			DSN:             dsn,
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: connLifetime,
		},
		JWT: JWTConfig{
			Secret:        secret,
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: viper.GetInt64("RATE_LIMIT_RPS"),
			Burst:             viper.GetInt64("RATE_LIMIT_BURST"),
		},
	}, nil
}
