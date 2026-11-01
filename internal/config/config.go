package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	SQLServerURL string
	JWTSecret    string
	TokenTTL     time.Duration
	DevTokens    bool
}

func Load() Config {
	return Config{
		Port:         env("PORT", "8080"),
		SQLServerURL: env("SQLSERVER_URL", "sqlserver://sa:Pivote_2026!@localhost:1433?database=pivote"),
		JWTSecret:    env("JWT_SECRET", "cambiar-en-produccion"),
		TokenTTL:     envDuration("TOKEN_TTL", time.Hour),
		DevTokens:    envBool("DEV_TOKENS", true),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	v, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}

func envBool(key string, def bool) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}
